import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from './auth'

export const useFriendsStore = defineStore('friends', () => {
  const friends = ref([])
  const pendingRequests = ref([])
  const activeTab = ref('messages') // 'messages' | 'pending' | 'add'
  const loading = ref(false)
  const error = ref(null)

  const pendingCount = computed(() => pendingRequests.value.length)

  function getHeaders() {
    const authStore = useAuthStore()
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authStore.token}`
    }
  }

  async function fetchFriends() {
    loading.value = true
    error.value = null
    try {
      const response = await fetch('/api/friends', {
        headers: getHeaders()
      })
      if (response.ok) {
        const data = await response.json()
        friends.value = Array.isArray(data) ? data : (data.friends || [])
      }
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  async function fetchPendingRequests() {
    loading.value = true
    error.value = null
    try {
      const response = await fetch('/api/friends/pending', {
        headers: getHeaders()
      })
      if (response.ok) {
        const data = await response.json()
        pendingRequests.value = Array.isArray(data) ? data : (data.requests || [])
      }
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  async function sendFriendRequest(target) {
    loading.value = true
    error.value = null
    try {
      let bodyData
      if (typeof target === 'number') {
        bodyData = { to_user_id: target }
      } else if (typeof target === 'string' && /^\d+$/.test(target.trim())) {
        bodyData = { to_user_id: parseInt(target.trim(), 10) }
      } else {
        bodyData = { to_username: String(target || '').trim() }
      }
      const response = await fetch('/api/friends/request', {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify(bodyData)
      })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || data.message || 'Failed to send request')
      }
      return data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function acceptFriendRequest(requestId) {
    loading.value = true
    error.value = null
    try {
      const response = await fetch('/api/friends/accept', {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ request_id: requestId })
      })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || data.message || 'Failed to accept request')
      }
      // Refresh list
      await fetchFriends()
      await fetchPendingRequests()
      return data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function searchUsers(query) {
    if (!query || !query.trim()) return []
    try {
      const response = await fetch(`/api/friends/search?q=${encodeURIComponent(query.trim())}`, {
        headers: getHeaders()
      })
      if (response.ok) {
        const data = await response.json()
        return Array.isArray(data) ? data : (data.users || [])
      }
      return []
    } catch (err) {
      console.error('Search users error:', err)
      return []
    }
  }

  return {
    friends,
    pendingRequests,
    activeTab,
    loading,
    error,
    pendingCount,
    fetchFriends,
    fetchPendingRequests,
    sendFriendRequest,
    acceptFriendRequest,
    searchUsers
  }
})

