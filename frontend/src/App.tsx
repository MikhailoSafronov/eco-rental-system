import { Routes, Route } from 'react-router-dom'
import Login from './pages/Login'
import Register from './pages/Register'
import Home from './pages/Home'
import { ProtectedRoute } from './components/ProtectedRoute'
import { Layout } from './components/Layout'
import Profile from './pages/Profile'
import Admin from './pages/Admin'

function App() {
  return (
    <div className="min-h-screen font-sans text-gray-900">
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        
        {/* Захищені маршрути */}
        <Route element={<ProtectedRoute />}>
          <Route element={<Layout />}>
            <Route path="/" element={<Home />} />
            <Route path="/profile" element={<Profile />} />
            <Route path="/admin" element={<Admin />} />
            <Route path="*" element={<div className="p-8 text-center text-red-500">404 - Сторінку не знайдено 😢</div>} />
          </Route>
        </Route>
      </Routes>
    </div>
  )
}

export default App
