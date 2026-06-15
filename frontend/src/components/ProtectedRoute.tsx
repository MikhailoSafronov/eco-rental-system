import { Navigate, Outlet } from 'react-router-dom';
import { useAuthStore } from '../store/useAuthStore';

export const ProtectedRoute = () => {
  // Отримуємо токен зі сховища Zustand
  const token = useAuthStore((state) => state.token);
  
  // Якщо токен є — користувач авторизований (true), якщо ні — гість (false)
  const isAuthenticated = !!token;

  if (!isAuthenticated) {
    // Якщо не авторизований — перенаправляємо на сторінку входу
    return <Navigate to="/login" replace />;
  }

  // Якщо все ок — пропускаємо до дочірніх маршрутів (Outlet)
  return <Outlet />;
};