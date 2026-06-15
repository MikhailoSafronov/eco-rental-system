import React, { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { api } from '../api/axios'
import { isAxiosError } from 'axios'

export default function Register() {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setIsLoading(true)

    try {
      // Відправляємо запит на бекенд для створення користувача
      await api.post('/users/register', { name, email, phone, password })
      
      // Якщо успіх - перекидаємо на сторінку входу
      // (Можна також додати глобальне повідомлення про успіх)
      navigate('/login')
      
    } catch (err) {
      if (isAxiosError(err)) {
        setError(err.response?.data?.error || 'Помилка реєстрації')
      } else {
        setError('Невідома помилка')
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-100">
      <div className="w-full max-w-md rounded-lg bg-white p-8 shadow-md">
        <h2 className="mb-6 text-center text-2xl font-bold text-gray-800">Створення акаунту 📝</h2>
        {error && <p className="mb-4 text-center text-sm text-red-500">{error}</p>}
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">Ім'я</label>
            <input type="text" value={name} onChange={(e) => setName(e.target.value)} required
              className="w-full rounded-md border border-gray-300 px-3 py-2 focus:border-green-500 focus:outline-none focus:ring-1 focus:ring-green-500" />
          </div>
          <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">Email</label>
            <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} required
              className="w-full rounded-md border border-gray-300 px-3 py-2 focus:border-green-500 focus:outline-none focus:ring-1 focus:ring-green-500" />
          </div>
          <div>
          <label className="mb-1 block text-sm font-medium text-gray-700">Телефон</label>
          <input type="tel" value={phone} onChange={(e) => setPhone(e.target.value)} required placeholder="+380..."
            className="w-full rounded-md border border-gray-300 px-3 py-2 focus:border-green-500 focus:outline-none focus:ring-1 focus:ring-green-500" />
        </div>
        <div>
            <label className="mb-1 block text-sm font-medium text-gray-700">Пароль</label>
            <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={6}
              className="w-full rounded-md border border-gray-300 px-3 py-2 focus:border-green-500 focus:outline-none focus:ring-1 focus:ring-green-500" />
          </div>
          
          <button type="submit" disabled={isLoading} className={`w-full rounded-md px-4 py-2 text-white transition ${isLoading ? 'bg-green-400 cursor-not-allowed' : 'bg-green-600 hover:bg-green-700'}`}>
            {isLoading ? 'Завантаження...' : 'Зареєструватися'}
          </button>
        </form>

        <p className="mt-6 text-center text-sm text-gray-600">
          Вже є акаунт?{' '}
          <Link to="/login" className="font-semibold text-green-600 hover:underline">Увійти</Link>
        </p>
      </div>
    </div>
  )
}