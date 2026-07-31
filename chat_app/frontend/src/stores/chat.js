import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth'
import { useFriendsStore } from './friends'
import { useRoomsStore } from './rooms'

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
  const unreadRoomCounts = ref({})
  const userStatus = ref({}) // key: user_id, value: { is_online, last_seen }

  const selectedRoomId = ref(null)
  const roomMessages = ref({}) // key: room_id, value: array of message objects

  const activeMessages = computed(() => {
    if (selectedRoomId.value) {
      return roomMessages.value[selectedRoomId.value] || []
    }
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
    selectedRoomId.value = null
    if (friendId) {
      if (unreadCounts.value[friendId] > 0) {
        unreadCounts.value[friendId] = 0
        sendReadReceipt(friendId)
      }
      fetchMessages(friendId)
    }
  }

  function selectRoom(roomId) {
    selectedRoomId.value = roomId
    selectedFriendId.value = null
    if (roomId) {
      fetchRoomMessages(roomId)
    }
  }

  async function fetchUnreadCounts() {
    try {
      const response = await fetch('/api/messages/unread', { headers: getHeaders() })
      if (response.ok) {
        const data = await response.json()
        unreadCounts.value = data.users || {}
        unreadRoomCounts.value = data.rooms || {}
      }
    } catch (err) {
      console.error('Failed to fetch unread counts', err)
    }
  }

  async function fetchMessages(friendId, beforeId = 0) {
    if (!friendId) return
    
    // Use cached messages if available for initial load
    if (beforeId === 0 && messages.value[friendId] && messages.value[friendId].length > 0) {
      return
    }

    loading.value = true
    error.value = null
    try {
      const url = beforeId > 0 
        ? `/api/messages/${friendId}?before_id=${beforeId}&limit=50`
        : `/api/messages/${friendId}?limit=50`

      const response = await fetch(url, {
        headers: getHeaders()
      })
      if (response.ok) {
        const data = await response.json()
        const newMessages = Array.isArray(data) ? data : []
        
        if (!messages.value[friendId]) {
          messages.value[friendId] = []
        }

        if (beforeId > 0) {
          // Prepend older messages
          messages.value[friendId] = [...newMessages, ...messages.value[friendId]]
        } else {
          // First load
          messages.value[friendId] = newMessages
        }
      }
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  async function fetchRoomMessages(roomId, beforeId = 0) {
    if (!roomId) return
    
    if (beforeId === 0 && roomMessages.value[roomId] && roomMessages.value[roomId].length > 0) {
      return
    }

    loading.value = true
    error.value = null
    try {
      const url = beforeId > 0 
        ? `/api/messages/room/${roomId}?before_id=${beforeId}&limit=50`
        : `/api/messages/room/${roomId}?limit=50`

      const response = await fetch(url, {
        headers: getHeaders()
      })
      if (response.ok) {
        const data = await response.json()
        const newMessages = Array.isArray(data) ? data : []
        
        if (!roomMessages.value[roomId]) {
          roomMessages.value[roomId] = []
        }

        if (beforeId > 0) {
          roomMessages.value[roomId] = [...newMessages, ...roomMessages.value[roomId]]
        } else {
          roomMessages.value[roomId] = newMessages
        }

        if (roomMessages.value[roomId] && roomMessages.value[roomId].length > 0) {
          const latestMsg = roomMessages.value[roomId][roomMessages.value[roomId].length - 1]
          markRoomAsRead(roomId, latestMsg.id)
        }
      }
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  function sendReadReceipt(senderId) {
    if (ws.value && ws.value.readyState === WebSocket.OPEN && senderId) {
      ws.value.send(JSON.stringify({
        type: 'read',
        receiver_id: senderId // Here receiver_id means the original sender whose messages we read
      }))
    } else {
      // Fallback to HTTP
      fetch(`/api/messages/read/${senderId}`, { method: 'POST', headers: getHeaders() }).catch(() => {})
    }
  }

  function appendMessage(msg) {
    const authStore = useAuthStore()
    const currentUserId = authStore.user?.id

    if (msg.room_id) {
      if (!roomMessages.value[msg.room_id]) {
        roomMessages.value[msg.room_id] = []
      }
      const list = roomMessages.value[msg.room_id]
      if (msg.id && list.some(m => m.id === msg.id)) {
        return
      }
      if (msg.id && msg.sender_id === currentUserId) {
        const tempIndex = list.findIndex(m => m.pending || (typeof m.id === 'string' && m.id.startsWith('temp-')))
        if (tempIndex !== -1 && list[tempIndex].content === msg.content) {
          list[tempIndex] = msg
          return
        }
      }
      list.push(msg)
      return
    }

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

  async function sendMessage(receiverId, content, msgType = 'text', fileUrl = '', roomId = null, replyToId = null) {
    if ((!content || !content.trim()) && !fileUrl) return
    const authStore = useAuthStore()
    const trimmed = content.trim()

    const tempMsg = {
      id: `temp-${Date.now()}`,
      sender_id: authStore.user?.id,
      receiver_id: receiverId,
      room_id: roomId,
      reply_to_id: replyToId,
      content: trimmed,
      type: msgType,
      file_url: fileUrl,
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
        room_id: roomId,
        reply_to_id: replyToId,
        content: trimmed,
        msg_type: msgType,
        file_url: fileUrl
      }))
    }
  }

  function editMessage(msgId, receiverId, newContent) {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({
        type: 'edit_message',
        id: msgId,
        receiver_id: receiverId,
        content: newContent
      }))
    }
  }

  function deleteMessage(msgId, receiverId, roomId = null) {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({
        type: 'delete_message',
        id: msgId,
        receiver_id: receiverId,
        room_id: roomId
      }))
    }
  }

  function markRoomAsRead(roomId, lastMessageId) {
    if (unreadRoomCounts.value[roomId] > 0) {
      unreadRoomCounts.value[roomId] = 0
    }
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({
        type: 'read_room',
        room_id: roomId,
        id: lastMessageId
      }))
    }
  }

  let reconnectAttempts = 0
  let reconnectTimer = null

  function connectWebSocket(token) {
    if (!token || (ws.value && (ws.value.readyState === WebSocket.OPEN || ws.value.readyState === WebSocket.CONNECTING))) {
      return
    }

    // Use environment variable if available, otherwise construct from current host
    let wsUrl = import.meta.env.VITE_WS_URL 
    if (!wsUrl) {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      let host = window.location.host
      if (host.includes(':5173')) {
        host = host.replace(':5173', ':3000')
      }
      wsUrl = `${protocol}//${host}`
    }
    wsUrl = `${wsUrl}/ws?token=${encodeURIComponent(token)}`

    try {
      const socket = new WebSocket(wsUrl)

      socket.onopen = () => {
        isConnected.value = true
        reconnectAttempts = 0
        if (reconnectTimer) {
          clearTimeout(reconnectTimer)
          reconnectTimer = null
        }
        console.log('[WS] Connected to', wsUrl)
      }

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          const friendsStore = useFriendsStore()
          const authStore = useAuthStore()

          if (data.type === 'chat') {
            if (data.sender_id && typingTimers[data.sender_id]) {
              clearTimeout(typingTimers[data.sender_id])
              delete typingTimers[data.sender_id]
              typingUsers.value[data.sender_id] = false
            }
            appendMessage(data)

            if (data.sender_id !== authStore.user?.id) {
              if (data.room_id) {
                if (data.room_id !== selectedRoomId.value) {
                  unreadRoomCounts.value[data.room_id] = (unreadRoomCounts.value[data.room_id] || 0) + 1
                } else {
                  markRoomAsRead(data.room_id, data.id)
                }
              } else if (data.sender_id !== selectedFriendId.value) {
                unreadCounts.value[data.sender_id] = (unreadCounts.value[data.sender_id] || 0) + 1
              }
              // Play notification sound
              try {
                const audio = new Audio('https://assets.mixkit.co/active_storage/sfx/2869/2869-preview.mp3')
                audio.play().catch(() => {})
              } catch(e) {}
            }
          } else if (data.type === 'typing') {
            handleIncomingTyping(data)
          } else if (data.type === 'read_receipt') {
            const readerId = data.reader_id
            if (messages.value[readerId]) {
              messages.value[readerId].forEach(m => {
                if (m.sender_id === authStore.user?.id) {
                  m.is_read = true
                }
              })
            }
          } else if (data.type === 'message_edited') {
            const listId = data.sender_id === authStore.user?.id ? data.receiver_id : data.sender_id
            if (messages.value[listId]) {
              const msgIndex = messages.value[listId].findIndex(m => m.id === data.id)
              if (msgIndex !== -1) {
                messages.value[listId][msgIndex].content = data.content
                messages.value[listId][msgIndex].is_edited = data.is_edited
              }
            }
          } else if (data.type === 'message_deleted') {
            // Find message in friend or room messages
            const lists = [messages.value, roomMessages.value]
            lists.forEach(map => {
              Object.keys(map).forEach(listId => {
                const msgIndex = map[listId].findIndex(m => m.id === data.id)
                if (msgIndex !== -1) {
                  map[listId][msgIndex].is_deleted = true
                }
              })
            })
          } else if (data.type === 'user_status') {
            userStatus.value[data.user_id] = {
              is_online: data.is_online,
              last_seen: data.last_seen
            }
          } else if (data.type === 'new_friend_request') {
            friendsStore.fetchPendingRequests()
          } else if (data.type === 'friend_request_accepted') {
            friendsStore.fetchFriends()
            friendsStore.fetchPendingRequests()
          } else if (data.type === 'room_invite') {
            const roomsStore = useRoomsStore()
            roomsStore.fetchPendingInvites()
          } else if (data.type === 'room_left' || data.type === 'room_kicked') {
            const roomsStore = useRoomsStore()
            roomsStore.rooms = roomsStore.rooms.filter(r => r.id != data.room_id)
            if (selectedRoomId.value == data.room_id) {
              selectedRoomId.value = null
              if (roomMessages.value[data.room_id]) {
                delete roomMessages.value[data.room_id]
              }
            }
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
        
        // Exponential backoff for reconnect
        const backoffMs = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000)
        reconnectAttempts++
        console.log(`[WS] Reconnecting in ${backoffMs}ms... (Attempt ${reconnectAttempts})`)
        reconnectTimer = setTimeout(() => {
          if (authStore?.token) connectWebSocket(authStore.token)
          else connectWebSocket(token)
        }, backoffMs)
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
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
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
    unreadCounts,
    unreadRoomCounts,
    fetchUnreadCounts,
    sendReadReceipt,
    editMessage,
    deleteMessage,
    markRoomAsRead,
    userStatus,
    selectedRoomId,
    roomMessages,
    selectRoom,
    fetchRoomMessages
  }
})

