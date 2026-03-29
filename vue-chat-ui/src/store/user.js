import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '../api'
import { normalizeUserAvatar } from '../utils/mediaUrl'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  // Try to load user info from localStorage to avoid flash of unauthenticated content
  const userInfo = ref(normalizeUserAvatar(JSON.parse(localStorage.getItem('user_info') || 'null')))

  function setToken(newToken) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function setUserInfo(info) {
    const normalized = normalizeUserAvatar(info)
    userInfo.value = normalized
    if (info) {
      localStorage.setItem('user_info', JSON.stringify(normalized))
    } else {
      localStorage.removeItem('user_info')
    }
  }

  async function getUserInfo() {
    if (!token.value) return
    try {
      const res = await api.get('/api/user/profile')
      if (res.data && res.data.user) {
        setUserInfo(res.data.user) // Use setter to persist
      }
    } catch (e) {
      console.error('Failed to get user info:', e)
      // Optional: logout if token invalid
    }
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user_info')
  }

  return { token, userInfo, setToken, setUserInfo, getUserInfo, logout }
})
