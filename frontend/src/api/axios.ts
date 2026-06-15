import axios from 'axios'
import { useAuthStore } from '../store/useAuthStore'

export const api = axios.create({
  // Звертаємося до нашого локального Go-сервера
  baseURL: 'http://localhost:8080/api',
  headers: { 'Content-Type': 'application/json' },
})

// Інтерсептор: перехоплює кожен запит ПЕРЕД відправкою на сервер
api.interceptors.request.use((config) => {
  // Дістаємо токен з нашого Zustand сховища
  const token = useAuthStore.getState().token
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Інтерсептор: перехоплює відповіді ВІД сервера
api.interceptors.response.use(
  (response) => response, // Якщо запит успішний - просто повертаємо дані
  (error) => {
    // Перевіряємо, чи це запит саме на авторизацію
    const isLoginRequest = error.config?.url?.includes('/users/login')

    // Якщо помилка 401 (Неавторизовано) і це НЕ спроба логіну
    if (error.response?.status === 401 && !isLoginRequest) {
      useAuthStore.getState().logout() // Очищаємо стан Zustand
      window.location.href = '/login' // Перекидаємо на сторінку логіну
    }
    return Promise.reject(error)
  }
)