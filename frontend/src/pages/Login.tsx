import React, { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { api } from '../api/axios'
import { useAuthStore } from '../store/useAuthStore'
import { isAxiosError } from 'axios'

export default function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  
  const navigate = useNavigate()
  const setAuth = useAuthStore((state) => state.setAuth)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    try {
      // Відправляємо запит на наш бекенд
      const response = await api.post('/users/login', { email, password })
      
      // Якщо успіх - зберігаємо токен і перекидаємо на головну
      setAuth(response.data.token, {
        id: response.data.user_id,
        name: '', email: '', role: 'client', balance: 0 // Деталі профілю завантажимо пізніше
      })
      navigate('/')
      
    } catch (err) {
      // Безпечна перевірка помилки через Axios
      if (isAxiosError(err)) {
        setError(err.response?.data?.error || 'Помилка авторизації')
      } else {
        setError('Невідома помилка')
      }
    }
  }

  return (
    <div className="flex h-screen items-center justify-center bg-gray-100">
      <div className="w-full max-w-md rounded-lg bg-white p-8 shadow-md">
        <h2 className="mb-6 text-center text-2xl font-bold text-gray-800">Вхід у систему</h2>
        {error && <p className="mb-4 text-center text-sm text-red-500">{error}</p>}
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">Email</label>
            <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required
              className="w-full rounded-md border border-gray-300 px-3 py-2 focus:border-green-500 focus:outline-none focus:ring-1 focus:ring-green-500" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">Пароль</label>
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required
              className="w-full rounded-md border border-gray-300 px-3 py-2 focus:border-green-500 focus:outline-none focus:ring-1 focus:ring-green-500" />
          </div>
          
          <button type="submit" className="w-full rounded-md bg-green-600 px-4 py-2 text-white transition hover:bg-green-700">
            Увійти
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-gray-600">
          Ще немає акаунту?{' '}
          <Link to="/register" className="font-semibold text-green-600 hover:underline">Зареєструватися</Link>
        </p>
      </div>
    </div>
  )
}
