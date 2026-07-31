import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useAuthStore } from './auth'

export const useRoomsStore = defineStore('rooms', () => {
  const rooms = ref([])
  const pendingInvites = ref([])
  const loading = ref(false)
  const error = ref(null)

  function getHeaders() {
    const authStore = useAuthStore()
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authStore.token}`
    }
  }

  async function fetchRooms() {
    loading.value = true
    error.value = null
    try {
      const response = await fetch('/api/rooms', { headers: getHeaders() })
      if (!response.ok) throw new Error('Failed to fetch rooms')
      rooms.value = await response.json() || []
    } catch (err) {
      error.value = err.message
    } finally {
      loading.value = false
    }
  }

  async function fetchPendingInvites() {
    try {
      const response = await fetch('/api/rooms/invites', { headers: getHeaders() })
      if (!response.ok) throw new Error('Failed to fetch invites')
      pendingInvites.value = await response.json() || []
    } catch (err) {
      console.error(err)
    }
  }

  async function createRoom(name, avatarUrl = '') {
    loading.value = true
    error.value = null
    try {
      const response = await fetch('/api/rooms', {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ name, avatar_url: avatarUrl })
      })
      const data = await response.json()
      if (!response.ok) throw new Error(data.error || 'Failed to create room')
      
      // Fetch rooms again to update the list
      await fetchRooms()
      return data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function inviteToRoom(roomId, toUserId) {
    loading.value = true
    error.value = null
    try {
      const response = await fetch(`/api/rooms/${roomId}/invite`, {
        method: 'POST',
        headers: getHeaders(),
        body: JSON.stringify({ user_id: toUserId })
      })
      const data = await response.json()
      if (!response.ok) throw new Error(data.error || 'Failed to send invite')
      return data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function acceptInvite(roomId) {
    loading.value = true
    error.value = null
    try {
      const response = await fetch(`/api/rooms/invites/${roomId}/accept`, {
        method: 'POST',
        headers: getHeaders()
      })
      const data = await response.json()
      if (!response.ok) throw new Error(data.error || 'Failed to accept invite')
      
      // Update local state
      pendingInvites.value = pendingInvites.value.filter(inv => inv.room_id !== roomId)
      await fetchRooms()
      return data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    rooms,
    pendingInvites,
    loading,
    error,
    fetchRooms,
    fetchPendingInvites,
    createRoom,
    inviteToRoom,
    acceptInvite
  }
})
