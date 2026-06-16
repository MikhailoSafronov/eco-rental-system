import { useState } from 'react';
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

const vehicleTypeLabels: Record<string, string> = {
  scooter: '🛴 Самокат',
  bike: '🚲 Велосипед',
  moped: '🛵 Мопед',
};

export default function Admin() {
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();

  // Стан для вибраного виду транспорту (щоб фільтрувати моделі)
  const [selectedType, setSelectedType] = useState('scooter');

  // Стан для модального вікна додавання транспорту
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [newVehicleData, setNewVehicleData] = useState<{
    model_id: number;
    tariff_id: number;
    latitude: number | string;
    longitude: number | string;
    battery_level: number | string;
  }>({
    model_id: 1,
    tariff_id: 1,
    latitude: 46.6322,
    longitude: 32.6146,
    battery_level: 100,
  });

  // Стан для модалки додавання зони паркування
  const [isZoneModalOpen, setIsZoneModalOpen] = useState(false);
  const [zoneName, setZoneName] = useState('');
  const [zoneCoordinates, setZoneCoordinates] = useState('');

  // Отримуємо ВСІ самокати (включно зі зламаними та орендованими)
  const { data: vehicles, isLoading, isError } = useQuery<AdminVehicle[]>({
    queryKey: ['admin', 'vehicles'],
    queryFn: async () => {
      const res = await api.get('/admin/vehicles');
      return res.data?.vehicles || res.data || [];
    }
  });

  // Отримуємо список всіх моделей
  const { data: models } = useQuery<{id: number, name: string, type: string}[]>({
    queryKey: ['admin', 'models'],
    queryFn: async () => {
      const res = await api.get('/admin/models');
      return res.data || [];
    }
  });

  // Отримуємо список всіх тарифів
  const { data: tariffs } = useQuery<{id: number, name: string, unlock_price: number, minute_price: number}[]>({
    queryKey: ['admin', 'tariffs'],
    queryFn: async () => {
      const res = await api.get('/admin/tariffs');
      return res.data || [];
    }
  });

  // Отримуємо список всіх паркувальних зон (для виводу і видалення)
  const { data: parkingZones } = useQuery<{id: number, name: string}[]>({
    queryKey: ['zones'],
    queryFn: async () => {
      const res = await api.get('/zones');
      return res.data?.zones || res.data || [];
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

  // Мутація для створення нового транспорту
  const addVehicleMutation = useMutation({
    mutationFn: async (data: typeof newVehicleData) => {
      // Перетворюємо координати на числа (відкидаючи можливі пусті рядки)
      const payload = {
        ...data,
        latitude: Number(data.latitude) || 0,
        longitude: Number(data.longitude) || 0,
        battery_level: data.battery_level === '' ? 100 : Number(data.battery_level),
      };
      const res = await api.post('/admin/vehicles', payload);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'vehicles'] });
      setIsAddModalOpen(false); // Закриваємо модалку при успіху
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Помилка';
      alert(`Помилка додавання транспорту:\n${message}`);
    }
  });

  // Мутація для створення зони
  const addZoneMutation = useMutation({
    mutationFn: async () => {
      // Парсимо введені координати (розбиваємо по рядках, чистимо від ком і літер)
      const lines = zoneCoordinates.split('\n').map(l => l.trim()).filter(l => l.length > 0);
      const points: [number, number][] = [];
      
      for (const line of lines) {
        const parts = line.replace(/,/g, ' ').replace(/[^\d.\s-]/g, '').trim().split(/\s+/);
        if (parts.length >= 2) {
          const lat = parseFloat(parts[0]);
          const lon = parseFloat(parts[1]);
          if (!isNaN(lat) && !isNaN(lon)) {
            points.push([lon, lat]); // PostGIS очікує формат [Довгота, Широта]
          }
        }
      }

      if (points.length < 3) {
        throw new Error('Необхідно ввести мінімум 3 коректні точки (Широта, Довгота)');
      }

      const res = await api.post('/admin/zones', { name: zoneName, points });
      return res.data;
    },
    onSuccess: () => {
      setIsZoneModalOpen(false);
      setZoneName('');
      setZoneCoordinates('');
      alert('Паркувальну зону успішно створено! 🗺️');
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : error.message || 'Помилка';
      alert(`Помилка додавання зони:\n${message}`);
    }
  });

  // Мутація для видалення транспорту
  const deleteVehicleMutation = useMutation({
    mutationFn: async (id: number) => {
      const res = await api.delete(`/admin/vehicles/${id}`);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'vehicles'] });
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Помилка';
      alert(`Помилка видалення:\n${message}`);
    }
  });

  // Мутація для видалення зони
  const deleteZoneMutation = useMutation({
    mutationFn: async (id: number) => {
      const res = await api.delete(`/admin/zones/${id}`);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['zones'] });
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Помилка';
      alert(`Помилка видалення зони:\n${message}`);
    }
  });

  // Перевірка ролі (захист на рівні фронтенду) ПОВИННА БУТИ ПІСЛЯ ВСІХ ХУКІВ
  if (user?.role !== 'admin') {
    return <Navigate to="/" replace />;
  }

  if (isLoading) return <div className="p-8 text-center text-gray-500">Завантаження автопарку...</div>;
  if (isError) return <div className="p-8 text-center text-red-500 font-bold">Помилка доступу до бази даних.</div>;

  return (
    <div className="mx-auto max-w-6xl p-6">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-3xl font-extrabold text-gray-800">Управління автопарком 🛠️</h1>
        <div className="flex gap-3">
          <button 
            onClick={() => setIsZoneModalOpen(true)}
            className="rounded-xl bg-green-600 px-4 py-2 text-white font-bold transition hover:bg-green-700 shadow-md"
          >
            + Додати зону
          </button>
          <button 
            onClick={() => setIsAddModalOpen(true)}
            className="rounded-xl bg-gray-800 px-4 py-2 text-white font-bold transition hover:bg-gray-700 shadow-md"
          >
            + Додати транспорт
          </button>
        </div>
      </div>

      <div className="overflow-hidden rounded-xl bg-white shadow-md">
        <table className="w-full text-left text-sm text-gray-600">
          <thead className="bg-gray-50 text-xs uppercase text-gray-700 border-b border-gray-200">
            <tr>
              <th className="px-6 py-4">ID / UUID</th>
              <th className="px-6 py-4">Модель</th>
              <th className="px-6 py-4">Заряд</th>
              <th className="px-6 py-4">Статус</th>
              <th className="px-6 py-4 text-right">Дії</th>
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
                  <div className="text-xs text-gray-500">
                    {v.vehicle_type ? vehicleTypeLabels[v.vehicle_type] || `❓ Інше (${v.vehicle_type})` : '❓ Не вказано'}
                  </div>
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
                <td className="px-6 py-4 text-right">
                  <button
                    onClick={() => {
                      if (window.confirm(`Ви впевнені, що хочете видалити транспорт #${v.id}?`)) {
                        deleteVehicleMutation.mutate(v.id);
                      }
                    }}
                    disabled={deleteVehicleMutation.isPending}
                    className="text-red-500 hover:text-red-700 font-bold transition disabled:opacity-50"
                  >
                    Видалити
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="mt-12 mb-6 flex items-center justify-between">
        <h2 className="text-2xl font-extrabold text-gray-800">Паркувальні зони 🗺️</h2>
      </div>
      
      <div className="overflow-hidden rounded-xl bg-white shadow-md mb-12">
        <table className="w-full text-left text-sm text-gray-600">
          <thead className="bg-gray-50 text-xs uppercase text-gray-700 border-b border-gray-200">
            <tr>
              <th className="px-6 py-4">ID</th>
              <th className="px-6 py-4">Назва зони</th>
              <th className="px-6 py-4 text-right">Дії</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {parkingZones?.map((zone) => (
              <tr key={zone.id} className="hover:bg-gray-50 transition">
                <td className="px-6 py-4 font-bold text-gray-800">#{zone.id}</td>
                <td className="px-6 py-4 font-medium text-gray-800">{zone.name}</td>
                <td className="px-6 py-4 text-right">
                  <button 
                    onClick={() => { if (window.confirm(`Видалити зону "${zone.name}"?`)) deleteZoneMutation.mutate(zone.id); }}
                    className="text-red-500 hover:text-red-700 font-bold transition"
                  >
                    Видалити
                  </button>
                </td>
              </tr>
            ))}
            {(!parkingZones || parkingZones.length === 0) && (
              <tr><td colSpan={3} className="px-6 py-4 text-center">Немає паркувальних зон</td></tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Модальне вікно для додавання транспорту */}
      {isAddModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h2 className="mb-4 text-xl font-bold text-gray-800">Додати новий транспорт 🛴</h2>
            
            <div className="flex flex-col gap-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Вид транспорту</label>
                <select 
                  value={selectedType} 
                  onChange={(e) => {
                    const newType = e.target.value;
                    setSelectedType(newType);
                    // Автоматично обираємо першу модель з нового виду
                    const firstModel = models?.find(m => m.type === newType);
                    if (firstModel) setNewVehicleData(prev => ({ ...prev, model_id: firstModel.id }));
                  }}
                  className="w-full cursor-pointer rounded-lg border border-gray-300 bg-white px-3 py-2 outline-none focus:border-gray-800"
                >
                  <option value="scooter">🛴 Самокат</option>
                  <option value="bike">🚲 Велосипед</option>
                  <option value="moped">🛵 Мопед</option>
                </select>
              </div>
              
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Модель</label>
                <select 
                  value={newVehicleData.model_id} 
                  onChange={e => setNewVehicleData({...newVehicleData, model_id: parseInt(e.target.value)})} 
                  className="w-full cursor-pointer rounded-lg border border-gray-300 bg-white px-3 py-2 outline-none focus:border-gray-800"
                >
                  {models?.filter(m => m.type === selectedType).map(m => (
                    <option key={m.id} value={m.id}>{m.name}</option>
                  ))}
                  {models?.filter(m => m.type === selectedType).length === 0 && (
                    <option value={0} disabled>Немає моделей для цього типу</option>
                  )}
                </select>
              </div>
              
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Тариф</label>
                <select 
                  value={newVehicleData.tariff_id} 
                  onChange={e => setNewVehicleData({...newVehicleData, tariff_id: parseInt(e.target.value)})} 
                  className="w-full cursor-pointer rounded-lg border border-gray-300 bg-white px-3 py-2 outline-none focus:border-gray-800"
                >
                  {tariffs?.map(t => (
                    <option key={t.id} value={t.id}>{t.name} ({t.unlock_price}₴ + {t.minute_price}₴/хв)</option>
                  ))}
                </select>
              </div>
              
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Заряд батареї (%)</label>
                <input type="number" min="0" max="100" value={newVehicleData.battery_level} onChange={e => setNewVehicleData({...newVehicleData, battery_level: e.target.value === '' ? '' : parseInt(e.target.value)})} className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-gray-800" />
              </div>
              
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Широта (Latitude)</label>
                <input type="number" step="0.0001" value={newVehicleData.latitude} onChange={e => setNewVehicleData({...newVehicleData, latitude: e.target.value === '' ? '' : parseFloat(e.target.value)})} className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-gray-800" />
              </div>
              
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Довгота (Longitude)</label>
                <input type="number" step="0.0001" value={newVehicleData.longitude} onChange={e => setNewVehicleData({...newVehicleData, longitude: e.target.value === '' ? '' : parseFloat(e.target.value)})} className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-gray-800" />
              </div>
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button 
                onClick={() => setIsAddModalOpen(false)} 
                className="rounded-lg bg-gray-200 px-4 py-2 font-medium text-gray-700 transition hover:bg-gray-300"
              >
                Скасувати
              </button>
              <button 
                onClick={() => addVehicleMutation.mutate(newVehicleData)}
                disabled={addVehicleMutation.isPending}
                className="rounded-lg bg-green-500 px-4 py-2 font-medium text-white transition hover:bg-green-600 disabled:opacity-50"
              >
                {addVehicleMutation.isPending ? 'Збереження...' : 'Зберегти'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Модальне вікно для додавання зони паркування */}
      {isZoneModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-lg rounded-xl bg-white p-6 shadow-xl">
            <h2 className="mb-4 text-xl font-bold text-gray-800">Додати зону паркування 🗺️</h2>
            
            <div className="flex flex-col gap-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Назва зони</label>
                <input type="text" value={zoneName} onChange={e => setZoneName(e.target.value)} placeholder="Наприклад: ТРЦ Фабрика" className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-gray-800" />
              </div>
              
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Координати (Широта, Довгота)</label>
                <p className="text-xs text-gray-500 mb-2">Вводьте кожну точку з нового рядка. Можна скопіювати з Google Maps (наприклад: 46.6322, 32.6146).</p>
                <textarea 
                  value={zoneCoordinates} 
                  onChange={e => setZoneCoordinates(e.target.value)} 
                  rows={8}
                  placeholder="46.640, 32.605&#10;46.640, 32.622&#10;46.632, 32.625"
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 font-mono text-sm outline-none focus:border-gray-800 resize-y" 
                />
              </div>
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button onClick={() => setIsZoneModalOpen(false)} className="rounded-lg bg-gray-200 px-4 py-2 font-medium text-gray-700 transition hover:bg-gray-300">
                Скасувати
              </button>
              <button 
                onClick={() => addZoneMutation.mutate()}
                disabled={addZoneMutation.isPending || !zoneName.trim() || !zoneCoordinates.trim()}
                className="rounded-lg bg-green-600 px-4 py-2 font-medium text-white transition hover:bg-green-700 disabled:opacity-50"
              >
                {addZoneMutation.isPending ? 'Збереження...' : 'Зберегти зону'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}