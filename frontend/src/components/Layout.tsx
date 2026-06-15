import { Outlet, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/useAuthStore';

export const Layout = () => {
  const logout = useAuthStore((state) => state.logout);
  const user = useAuthStore((state) => state.user);
  const navigate = useNavigate();

  const handleLogout = () => {
    logout(); // Очищаємо токен та юзера в Zustand (і localStorage)
    navigate('/login'); // Перекидаємо на сторінку логіну
  };

  return (
    <div className="flex flex-col min-h-screen">
      {/* Це спільний хедер, який буде видно на всіх захищених сторінках */}
      <header className="flex items-center justify-between p-4 bg-green-600 text-white font-bold shadow-md">
        <span>Eco Rental - Оренда міського транспорту 🛴🚲🛵</span>
        <div className="flex items-center gap-4">
          {/* Показуємо email або ім'я користувача, якщо вони є */}
          {user && <span className="text-sm font-normal">Привіт, {user.email}!</span>}
          <button onClick={handleLogout} className="rounded bg-green-700 px-3 py-1 text-sm transition hover:bg-green-800">
            Вийти
          </button>
        </div>
      </header>
      
      <main className="flex-grow bg-gray-100">
        {/* Outlet — це місце, куди підставлятимуться сторінки (Головна, Профіль тощо) */}
        <Outlet />
      </main>
    </div>
  );
};