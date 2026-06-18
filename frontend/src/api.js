import axios from 'axios'

// Запросы идут на относительный /api: в dev их проксирует Vite, в prod — nginx.
const api = axios.create()

// Подставляем JWT в каждый запрос.
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// При 401 — разлогиниваем и уводим на страницу входа.
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('token')
      const path = window.location.pathname
      if (path !== '/login' && path !== '/register') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

// Достаёт человекочитаемый текст ошибки из ответа API.
export function apiError(error, fallback = 'Что-то пошло не так') {
  return error?.response?.data?.error || fallback
}

export default api
