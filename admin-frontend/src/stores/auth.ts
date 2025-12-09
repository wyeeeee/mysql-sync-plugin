import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('admin_token'))
  const expiresAt = ref<number>(Number(localStorage.getItem('admin_expires_at')) || 0)
  const username = ref<string>(localStorage.getItem('admin_username') || '')

  const isLoggedIn = computed(() => {
    if (!token.value) return false
    if (Date.now() > expiresAt.value * 1000) {
      logout()
      return false
    }
    return true
  })

  function setAuth(data: { token: string; expiresAt: number }, user: string) {
    token.value = data.token
    expiresAt.value = data.expiresAt
    username.value = user
    localStorage.setItem('admin_token', data.token)
    localStorage.setItem('admin_expires_at', String(data.expiresAt))
    localStorage.setItem('admin_username', user)
  }

  function logout() {
    token.value = null
    expiresAt.value = 0
    username.value = ''
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_expires_at')
    localStorage.removeItem('admin_username')
  }

  return {
    token,
    expiresAt,
    username,
    isLoggedIn,
    setAuth,
    logout
  }
})
