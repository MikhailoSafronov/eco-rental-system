import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/axios';
import { isAxiosError } from 'axios';
import { useAuthStore } from '../store/useAuthStore';
import { Navigate } from 'react-router-dom';

// Описуємо структуру транспорту для адмінки
interface AdminVehicle {
  id: number;
  uuid: string;
  model_name?: string;
  vehicle_type?: string;
  battery_level: number;
  status: string;
  latitude: number;
  longitude: number;
}

export default function Admin() {
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();

  // Перевірка ролі (захист на рівні фронтенду)
  if (user?.role !== 'admin') {
    return <Navigate to="/" replace />;
  }

  // Отримуємо ВСІ самокати (включно зі зламаними та орендованими)
  const { data: vehicles, isLoading, isError } = useQuery<AdminVehicle[]>({
    queryKey: ['admin', 'vehicles'],
    queryFn: async () => {
      const res = await api.get('/admin/vehicles');
      return res.data?.vehicles || res.data || [];
    }
  });

  // Мутація для зміни статусу (наприклад, відправити на ремонт)
  const updateStatusMutation = useMutation({
    mutationFn: async ({ id, status }: { id: number; status: string }) => {
      const res = await api.patch(`/admin/vehicles/${id}/status`, { status });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'vehicles'] });
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Помилка';
      alert(`Помилка оновлення статусу:\n${message}`);
    }
  });

  if (isLoading) return <div className="p-8 text-center text-gray-500">Завантаження автопарку...</div>;
  if (isError) return <div className="p-8 text-center text-red-500 font-bold">Помилка доступу до бази даних.</div>;

  return (
    <div className="mx-auto max-w-6xl p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-extrabold text-gray-800">Управління автопарком 🛠️</h1>
        <button className="rounded-xl bg-gray-800 px-4 py-2 text-white font-bold transition hover:bg-gray-700 shadow-md">
          + Додати транспорт
        </button>
      </div>

      <div className="overflow-hidden rounded-xl bg-white shadow-md">
        <table className="w-full text-left text-sm text-gray-600">
          <thead className="bg-gray-50 text-xs uppercase text-gray-700 border-b border-gray-200">
            <tr>
              <th className="px-6 py-4">ID / UUID</th>
              <th className="px-6 py-4">Модель</th>
              <th className="px-6 py-4">Заряд</th>
              <th className="px-6 py-4">Статус</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {vehicles?.map((v) => (
              <tr key={v.id} className="hover:bg-gray-50 transition">
                <td className="px-6 py-4">
                  <div className="font-bold text-gray-800">#{v.id}</div>
                  <div className="text-xs text-gray-400 font-mono">{v.uuid.split('-')[0]}...</div>
                </td>
                <td className="px-6 py-4">
                  <div className="font-medium text-gray-800">{v.model_name || 'Невідомо'}</div>
                  <div className="text-xs text-gray-500">{v.vehicle_type === 'scooter' ? '🛴 Самокат' : v.vehicle_type === 'bike' ? '🚲 Велосипед' : '🛵 Мопед'}</div>
                </td>
                <td className="px-6 py-4">
                  <span className={`font-bold ${v.battery_level < 20 ? 'text-red-500' : 'text-green-600'}`}>{v.battery_level}%</span>
                </td>
                <td className="px-6 py-4">
                  <select 
                    value={v.status}
                    onChange={(e) => updateStatusMutation.mutate({ id: v.id, status: e.target.value })}
                    disabled={updateStatusMutation.isPending}
                    className="cursor-pointer rounded-lg border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 shadow-sm outline-none transition focus:border-gray-800"
                  >
                    <option value="available">🟢 Доступний</option>
                    <option value="rented">🔵 В оренді</option>
                    <option value="low_battery">🟠 Розряджений</option>
                    <option value="maintenance">🛠 На ремонті</option>
                    <option value="broken">🔴 Зламаний</option>
                  </select>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}