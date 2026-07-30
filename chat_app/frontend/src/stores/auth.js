import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || null)
  const initialUser = (() => {
    try {
      const stored = localStorage.getItem('user')
      return stored ? JSON.parse(stored) : null
    } catch {
      localStorage.removeItem('user')
      return null
    }
  })()
  const user = ref(initialUser)
  const loading = ref(false)
  const error = ref(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  function setAuth(newToken, newUser) {
    token.value = newToken
    user.value = newUser
    if (newToken) {
      localStorage.setItem('token', newToken)
    } else {
      localStorage.removeItem('token')
    }
    if (newUser) {
      localStorage.setItem('user', JSON.stringify(newUser))
    } else {
      localStorage.removeItem('user')
    }
  }

  async function login(email, password) {
    loading.value = true
    error.value = null
    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password }),
      })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || data.message || 'Login failed')
      }
      setAuth(data.token, data.user)
      return data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function register(username, email, password, display_name, bio) {
    loading.value = true
    error.value = null
    try {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, email, password, display_name, bio }),
      })
      const data = await response.json()
      if (!response.ok) {
        throw new Error(data.error || data.message || 'Registration failed')
      }
      // do not call setAuth since we redirect to login
      return data
    } catch (err) {
      error.value = err.message
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchMe() {
    if (!token.value) return null
    loading.value = true
    try {
      const response = await fetch('/api/auth/me', {
        headers: {
          'Authorization': `Bearer ${token.value}`
        }
      })
      if (!response.ok) {
        logout()
        return null
      }
      const data = await response.json()
      if (data.user) {
        user.value = data.user
        localStorage.setItem('user', JSON.stringify(data.user))
      }
      return data.user
    } catch (err) {
      console.error('Fetch me error:', err)
      return null
    } finally {
      loading.value = false
    }
  }

  function logout() {
    setAuth(null, null)
  }

  function initAuth() {
    if (token.value && user.value) {
      fetchMe().catch(() => {})
    }
  }

  return {
    token,
    user,
    loading,
    error,
    isAuthenticated,
    login,
    register,
    fetchMe,
    logout,
    initAuth,
    setAuth
  }
})
