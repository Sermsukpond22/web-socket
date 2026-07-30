import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth'
import { useFriendsStore } from './friends'

export const useChatStore = defineStore('chat', () => {
  const selectedFriendId = ref(null)
  const messages = ref({}) // key: friend_id, value: array of message objects
  const ws = ref(null)
  const isConnected = ref(false)
  const loading = ref(false)
  const error = ref(null)
  const typingUsers = ref({})
  const typingTimers = {}
  const unreadCounts = ref({})

  const activeMessages = computed(() => {
    if (!selectedFriendId.value) return []
    return messages.value[selectedFriendId.value] || []
  })

  function isUserTyping(userId) {
    if (!userId) return false
    return !!typingUsers.value[userId]
  }

  function sendTyping(receiverId, isTyping = true) {
    if (ws.value && ws.value.readyState === WebSocket.OPEN && receiverId) {
      ws.value.send(JSON.stringify({
        type: 'typing',
        receiver_id: receiverId,
        is_typing: isTyping
      }))
    }
  }

  function handleIncomingTyping(data) {
    const senderId = data.sender_id
    if (!senderId) return

    if (data.is_typing === false) {
      if (typingTimers[senderId]) {
        clearTimeout(typingTimers[senderId])
        delete typingTimers[senderId]
      }
      typingUsers.value[senderId] = false
      return
    }

    typingUsers.value[senderId] = true

    if (typingTimers[senderId]) {
      clearTimeout(typingTimers[senderId])
    }

    typingTimers[senderId] = setTimeout(() => {
      typingUsers.value[senderId] = false
      delete typingTimers[senderId]
    }, 3000)
  }

  function getHeaders() {
    const authStore = useAuthStore()
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authStore.token}`
    }
  }

  function selectFriend(friendId) {
    selectedFriendId.value = friendId
    if (friendId) {
      unreadCounts.value[friendId] = 0
      fetchMessages(friendId)
    }
  }

  async function fetchMessages(friendId) {
    if (!friendId) return
    loading.value = true
    error.value = null
    try {
      const response = await fetch(`/api/messages/${friendId}`, {
        headers: getHeaders()
      })
      if (response.ok) {
        const data = await response.json()
        messages.value[friendId] = Array.isArray(data) ? data : []
      }
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  function appendMessage(msg) {
    const authStore = useAuthStore()
    const currentUserId = authStore.user?.id

    // Determine target conversation ID
    let conversationId
    if (msg.sender_id === currentUserId) {
      conversationId = msg.receiver_id
    } else {
      conversationId = msg.sender_id
    }

    if (!conversationId) return

    if (!messages.value[conversationId]) {
      messages.value[conversationId] = []
    }

    const list = messages.value[conversationId]

    // Check if exact message ID already exists
    if (msg.id && list.some(m => m.id === msg.id)) {
      return
    }

    // If incoming message has a server ID and is from current user, check if we have a pending temp message to replace
    if (msg.id && msg.sender_id === currentUserId) {
      const tempIndex = list.findIndex(m => m.pending || (typeof m.id === 'string' && m.id.startsWith('temp-')))
      if (tempIndex !== -1 && list[tempIndex].content === msg.content) {
        list[tempIndex] = msg
        return
      }
    }

    list.push(msg)
  }

  async function sendMessage(receiverId, content) {
    if (!content || !content.trim()) return
    const authStore = useAuthStore()
    const trimmed = content.trim()

    const tempMsg = {
      id: `temp-${Date.now()}`,
      sender_id: authStore.user?.id,
      receiver_id: receiverId,
      content: trimmed,
      created_at: new Date().toISOString(),
      pending: true
    }

    // Append to local store immediately for instant UX
    appendMessage(tempMsg)

    // Send over WebSocket if connected
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({
        type: 'chat',
        receiver_id: receiverId,
        content: trimmed
      }))
    }
  }

  function connectWebSocket(token) {
    if (!token || (ws.value && (ws.value.readyState === WebSocket.OPEN || ws.value.readyState === WebSocket.CONNECTING))) {
      return
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    let host = window.location.host
    if (host.includes(':5173')) {
      host = host.replace(':5173', ':3000')
    }
    const wsUrl = `${protocol}//${host}/ws?token=${encodeURIComponent(token)}`

    try {
      const socket = new WebSocket(wsUrl)

      socket.onopen = () => {
        isConnected.value = true
        console.log('[WS] Connected to', wsUrl)
      }

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          const friendsStore = useFriendsStore()

          if (data.type === 'chat') {
            if (data.sender_id && typingTimers[data.sender_id]) {
              clearTimeout(typingTimers[data.sender_id])
              delete typingTimers[data.sender_id]
              typingUsers.value[data.sender_id] = false
            }
            appendMessage(data)

            const authStore = useAuthStore()
            if (data.sender_id !== authStore.user?.id && data.sender_id !== selectedFriendId.value) {
              unreadCounts.value[data.sender_id] = (unreadCounts.value[data.sender_id] || 0) + 1
            }
          } else if (data.type === 'typing') {
            handleIncomingTyping(data)
          } else if (data.type === 'new_friend_request') {
            friendsStore.fetchPendingRequests()
          } else if (data.type === 'friend_request_accepted') {
            friendsStore.fetchFriends()
            friendsStore.fetchPendingRequests()
          } else if (data.type === 'error') {
            console.error('[WS Error]', data.message)
            error.value = data.message
          }
        } catch (e) {
          console.error('[WS Parse Error]', e)
        }
      }

      socket.onclose = () => {
        isConnected.value = false
        console.log('[WS] Disconnected')
      }

      socket.onerror = (err) => {
        isConnected.value = false
        console.error('[WS Socket Error]', err)
      }

      ws.value = socket
    } catch (e) {
      console.error('[WS Connection Exception]', e)
    }
  }

  function disconnectWebSocket() {
    if (ws.value) {
      ws.value.close()
      ws.value = null
      isConnected.value = false
    }
    Object.keys(typingTimers).forEach(id => clearTimeout(typingTimers[id]))
    typingUsers.value = {}
  }

  return {
    selectedFriendId,
    messages,
    ws,
    isConnected,
    loading,
    error,
    typingUsers,
    activeMessages,
    selectFriend,
    fetchMessages,
    sendMessage,
    appendMessage,
    connectWebSocket,
    disconnectWebSocket,
    sendTyping,
    isUserTyping,
    unreadCounts
  }
})

