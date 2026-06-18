import { defineStore } from 'pinia'
import api from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    user: null,
  }),
  getters: {
    isAuthenticated: (state) => !!state.token,
  },
  actions: {
    async login(email, password) {
      const { data } = await api.post('/api/auth/login', { email, password })
      this.setAuth(data)
    },
    async register(email, password) {
      const { data } = await api.post('/api/auth/register', { email, password })
      this.setAuth(data)
    },
    setAuth(data) {
      this.token = data.token
      this.user = data.user
      localStorage.setItem('token', data.token)
    },
    async fetchMe() {
      if (!this.token) return
      try {
        const { data } = await api.get('/api/me')
        this.user = data
      } catch {
        this.logout()
      }
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem('token')
    },
  },
})
