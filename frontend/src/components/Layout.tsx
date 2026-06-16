import { Outlet, useNavigate, Link } from 'react-router-dom';
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
        <Link to="/" className="hover:text-green-200 transition">
          Eco Rental - Оренда міського транспорту 🛴🚲🛵
        </Link>
        <div className="flex items-center gap-4">
          {/* Показуємо email або ім'я користувача, якщо вони є */}
          {user && <span className="text-sm font-normal">Привіт, {user.email}!</span>}
          {user?.role === 'admin' && (
            <Link to="/admin" className="rounded border border-gray-600 bg-gray-800 px-3 py-1 text-sm font-medium transition hover:bg-gray-700">
              Адмін-панель
            </Link>
          )}
          <Link to="/profile" className="rounded bg-green-500 px-3 py-1 text-sm font-medium transition hover:bg-green-400">
            Мій профіль
          </Link>
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