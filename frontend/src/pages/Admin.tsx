import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/axios';
import { isAxiosError } from 'axios';
import { useAuthStore } from '../store/useAuthStore';
import { Navigate } from 'react-router-dom';
import { MapContainer, TileLayer, Polyline } from 'react-leaflet';

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

// Описуємо структуру користувача
interface AdminUser {
  id: number;
  name: string;
  email: string;
  phone: string;
  role: string;
  balance: number;
  is_blocked: boolean;
}

// Описуємо структуру поїздки для адмінки
interface AdminRide {
  id: number;
  user_id: number;
  user_email: string;
  vehicle_id: number;
  vehicle_uuid: string;
  status: string;
  start_time: string;
  end_time?: string;
  total_price: number;
  track?: { latitude: number; longitude: number; timestamp: string }[];
}

const vehicleTypeLabels: Record<string, string> = {
  scooter: '🛴 Самокат',
  bike: '🚲 Велосипед',
  moped: '🛵 Мопед',
  monowheel: '🛞 Моноколесо',
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

  // Стани для редагування тарифів
  const [isEditTariffModalOpen, setIsEditTariffModalOpen] = useState(false);
  const [editingTariff, setEditingTariff] = useState<{ id: number, name: string, unlock_price: number, minute_price: number } | null>(null);
  const [isAddTariffModalOpen, setIsAddTariffModalOpen] = useState(false);
  const [newTariffData, setNewTariffData] = useState({ name: '', vehicle_type: 'scooter', unlock_price: 15, minute_price: 3.5 });

  // Стани для моделей транспорту
  const [isAddModelModalOpen, setIsAddModelModalOpen] = useState(false);
  const [newModelData, setNewModelData] = useState({ name: '', vehicle_type: 'scooter', battery_capacity_wh: 500, max_speed: 25 });
  
  // Стан для перегляду деталей поїздки (маршруту)
  const [selectedRide, setSelectedRide] = useState<AdminRide | null>(null);

  // Отримуємо ВСІ самокати (включно зі зламаними та орендованими)
  const { data: vehicles, isLoading, isError } = useQuery<AdminVehicle[]>({
    queryKey: ['admin', 'vehicles'],
    queryFn: async () => {
      const res = await api.get('/admin/vehicles');
      return res.data?.vehicles || res.data || [];
    }
  });

  // Отримуємо список всіх моделей
  const { data: models } = useQuery<{id: number, name: string, type: string, battery_capacity_wh: number, max_speed: number}[]>({
    queryKey: ['admin', 'models'],
    queryFn: async () => {
      const res = await api.get('/admin/models');
      return res.data || [];
    }
  });

  // Отримуємо список всіх тарифів
  const { data: tariffs } = useQuery<{id: number, name: string, vehicle_type: string, unlock_price: number, minute_price: number}[]>({
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

  // Отримуємо всіх користувачів
  const { data: usersData } = useQuery<AdminUser[]>({
    queryKey: ['admin', 'users'],
    queryFn: async () => {
      const res = await api.get('/admin/users');
      return res.data || [];
    }
  });

  // Отримуємо всі поїздки
  const { data: ridesData } = useQuery<AdminRide[]>({
    queryKey: ['admin', 'rides'],
    queryFn: async () => {
      const res = await api.get('/admin/rides');
      return res.data || [];
    }
  });

  // Отримуємо статистику для дашборду
  const { data: stats } = useQuery<{total_users: number, active_rides: number, total_revenue: number, total_vehicles: number}>({
    queryKey: ['admin', 'stats'],
    queryFn: async () => {
      const res = await api.get('/admin/stats');
      return res.data;
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

  // Мутація для оновлення тарифу
  const updateTariffMutation = useMutation({
    mutationFn: async (data: { id: number; unlock_price: number; minute_price: number }) => {
      const res = await api.patch(`/admin/tariffs/${data.id}`, {
        unlock_price: data.unlock_price,
        minute_price: data.minute_price,
      });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'tariffs'] });
      setIsEditTariffModalOpen(false);
      setEditingTariff(null);
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Помилка';
      alert(`Помилка оновлення тарифу:\n${message}`);
    }
  });

  // Мутація для створення тарифу
  const addTariffMutation = useMutation({
    mutationFn: async (data: typeof newTariffData) => {
      const res = await api.post('/admin/tariffs', data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'tariffs'] });
      setIsAddTariffModalOpen(false);
      setNewTariffData({ name: '', vehicle_type: 'scooter', unlock_price: 15, minute_price: 3.5 });
      alert('Новий тариф успішно створено! 💸');
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Помилка';
      alert(`Помилка додавання тарифу:\n${message}`);
    }
  });

  // Мутація для видалення тарифу
  const deleteTariffMutation = useMutation({
    mutationFn: async (id: number) => {
      const res = await api.delete(`/admin/tariffs/${id}`);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'tariffs'] });
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Помилка';
      alert(`Помилка видалення тарифу:\n${message}`);
    }
  });

  // Мутація для блокування/розблокування користувача
  const toggleUserBlockMutation = useMutation({
    mutationFn: async ({ id, is_blocked }: { id: number; is_blocked: boolean }) => {
      const res = await api.patch(`/admin/users/${id}/block`, { is_blocked });
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] });
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Помилка';
      alert(`Помилка зміни статусу користувача:\n${message}`);
    }
  });

  // Мутація для створення моделі
  const addModelMutation = useMutation({
    mutationFn: async (data: typeof newModelData) => {
      const res = await api.post('/admin/models', data);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'models'] });
      setIsAddModelModalOpen(false);
      setNewModelData({ name: '', vehicle_type: 'scooter', battery_capacity_wh: 500, max_speed: 25 });
      alert('Нову модель успішно створено! ⚙️');
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Помилка';
      alert(`Помилка додавання моделі:\n${message}`);
    }
  });

  // Мутація для видалення моделі
  const deleteModelMutation = useMutation({
    mutationFn: async (id: number) => {
      const res = await api.delete(`/admin/models/${id}`);
      return res.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'models'] });
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Помилка';
      alert(`Помилка видалення моделі:\n${message}`);
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

      {/* Дашборд статистики */}
      <div className="mb-10 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <div className="rounded-xl border border-gray-100 bg-white p-6 shadow-sm flex flex-col justify-center">
          <div className="text-sm font-medium text-gray-500">Загальний дохід</div>
          <div className="mt-2 text-3xl font-extrabold text-green-600">{stats?.total_revenue?.toFixed(2) || '0.00'} ₴</div>
        </div>
        <div className="rounded-xl border border-gray-100 bg-white p-6 shadow-sm flex flex-col justify-center">
          <div className="text-sm font-medium text-gray-500">Активні поїздки</div>
          <div className="mt-2 text-3xl font-extrabold text-blue-600">{stats?.active_rides || 0}</div>
        </div>
        <div className="rounded-xl border border-gray-100 bg-white p-6 shadow-sm flex flex-col justify-center">
          <div className="text-sm font-medium text-gray-500">Клієнтів у системі</div>
          <div className="mt-2 text-3xl font-extrabold text-gray-800">{stats?.total_users || 0}</div>
        </div>
        <div className="rounded-xl border border-gray-100 bg-white p-6 shadow-sm flex flex-col justify-center">
          <div className="text-sm font-medium text-gray-500">Транспорту в парку</div>
          <div className="mt-2 text-3xl font-extrabold text-gray-800">{stats?.total_vehicles || 0}</div>
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

      <div className="mt-12 mb-6 flex items-center justify-between">
        <h2 className="text-2xl font-extrabold text-gray-800">Управління моделями ⚙️</h2>
        <button 
          onClick={() => setIsAddModelModalOpen(true)}
          className="rounded-xl bg-gray-800 px-4 py-2 text-sm text-white font-bold transition hover:bg-gray-700 shadow-md"
        >
          + Додати модель
        </button>
      </div>
      
      <div className="overflow-hidden rounded-xl bg-white shadow-md mb-12">
        <table className="w-full text-left text-sm text-gray-600">
          <thead className="bg-gray-50 text-xs uppercase text-gray-700 border-b border-gray-200">
            <tr>
              <th className="px-6 py-4">ID</th>
              <th className="px-6 py-4">Назва моделі</th>
              <th className="px-6 py-4">Вид транспорту</th>
              <th className="px-6 py-4">Батарея (Вт·год)</th>
              <th className="px-6 py-4">Швидкість (км/год)</th>
              <th className="px-6 py-4 text-right">Дії</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {models?.map((model) => (
              <tr key={model.id} className="hover:bg-gray-50 transition">
                <td className="px-6 py-4 font-bold text-gray-800">#{model.id}</td>
                <td className="px-6 py-4 font-medium text-gray-800">{model.name}</td>
                <td className="px-6 py-4">{vehicleTypeLabels[model.type] || model.type}</td>
                <td className="px-6 py-4">{model.battery_capacity_wh}</td>
                <td className="px-6 py-4">{model.max_speed}</td>
                <td className="px-6 py-4 text-right">
                  <button 
                    onClick={() => { if (window.confirm(`Видалити модель "${model.name}"?`)) deleteModelMutation.mutate(model.id); }}
                    className="text-red-500 hover:text-red-700 font-bold transition"
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
        <h2 className="text-2xl font-extrabold text-gray-800">Управління тарифами 💰</h2>
        <button 
          onClick={() => setIsAddTariffModalOpen(true)}
          className="rounded-xl bg-green-600 px-4 py-2 text-sm text-white font-bold transition hover:bg-green-700 shadow-md"
        >
          + Додати тариф
        </button>
      </div>
      
      <div className="overflow-hidden rounded-xl bg-white shadow-md mb-12">
        <table className="w-full text-left text-sm text-gray-600">
          <thead className="bg-gray-50 text-xs uppercase text-gray-700 border-b border-gray-200">
            <tr>
              <th className="px-6 py-4">ID</th>
              <th className="px-6 py-4">Назва тарифу</th>
              <th className="px-6 py-4">Вид транспорту</th>
              <th className="px-6 py-4">Ціна розблокування (₴)</th>
              <th className="px-6 py-4">Ціна за хвилину (₴)</th>
              <th className="px-6 py-4 text-right">Дії</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {tariffs?.map((tariff) => (
              <tr key={tariff.id} className="hover:bg-gray-50 transition">
                <td className="px-6 py-4 font-bold text-gray-800">#{tariff.id}</td>
                <td className="px-6 py-4 font-medium text-gray-800">{tariff.name}</td>
                <td className="px-6 py-4">
                  {vehicleTypeLabels[tariff.vehicle_type] || tariff.vehicle_type}
                </td>
                <td className="px-6 py-4"><span className="font-bold text-green-600">{tariff.unlock_price} ₴</span></td>
                <td className="px-6 py-4"><span className="font-bold text-green-600">{tariff.minute_price} ₴</span></td>
                <td className="px-6 py-4 text-right">
                  <button 
                    onClick={() => { setEditingTariff(tariff); setIsEditTariffModalOpen(true); }}
                    className="text-blue-500 hover:text-blue-700 font-bold transition mr-4"
                  >
                    Редагувати
                  </button>
                  <button 
                    onClick={() => { if (window.confirm(`Видалити тариф "${tariff.name}"? Зауважте: неможливо видалити тариф, якщо він використовується транспортом.`)) deleteTariffMutation.mutate(tariff.id); }}
                    className="text-red-500 hover:text-red-700 font-bold transition"
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
        <h2 className="text-2xl font-extrabold text-gray-800">Керування користувачами 👥</h2>
      </div>
      
      <div className="overflow-hidden rounded-xl bg-white shadow-md mb-12">
        <table className="w-full text-left text-sm text-gray-600">
          <thead className="bg-gray-50 text-xs uppercase text-gray-700 border-b border-gray-200">
            <tr>
              <th className="px-6 py-4">ID</th>
              <th className="px-6 py-4">Користувач</th>
              <th className="px-6 py-4">Контакти</th>
              <th className="px-6 py-4">Баланс</th>
              <th className="px-6 py-4">Статус</th>
              <th className="px-6 py-4 text-right">Дії</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {usersData?.map((u) => (
              <tr key={u.id} className="hover:bg-gray-50 transition">
                <td className="px-6 py-4 font-bold text-gray-800">#{u.id}</td>
                <td className="px-6 py-4">
                  <div className="font-bold text-gray-800">{u.name}</div>
                  <div className="text-xs text-gray-500 uppercase">{u.role}</div>
                </td>
                <td className="px-6 py-4">
                  <div className="text-gray-800">{u.email}</div>
                  <div className="text-gray-500">{u.phone}</div>
                </td>
                <td className="px-6 py-4">
                  <span className={`font-bold ${u.balance < 0 ? 'text-red-500' : 'text-green-600'}`}>{u.balance.toFixed(2)} ₴</span>
                </td>
                <td className="px-6 py-4">
                  {u.is_blocked ? (
                    <span className="inline-flex items-center rounded-full bg-red-100 px-2.5 py-0.5 text-xs font-medium text-red-800">Заблокований</span>
                  ) : (
                    <span className="inline-flex items-center rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800">Активний</span>
                  )}
                </td>
                <td className="px-6 py-4 text-right">
                  {u.role !== 'admin' && (
                    <button 
                      onClick={() => {
                        if (window.confirm(`Ви впевнені, що хочете ${u.is_blocked ? 'РОЗБЛОКУВАТИ' : 'ЗАБЛОКУВАТИ'} користувача ${u.name}?`)) {
                          toggleUserBlockMutation.mutate({ id: u.id, is_blocked: !u.is_blocked });
                        }
                      }}
                      disabled={toggleUserBlockMutation.isPending}
                      className={`font-bold transition disabled:opacity-50 ${u.is_blocked ? 'text-green-600 hover:text-green-800' : 'text-red-500 hover:text-red-700'}`}
                    >
                      {u.is_blocked ? 'Розблокувати' : 'Заблокувати'}
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="mt-12 mb-6 flex items-center justify-between">
        <h2 className="text-2xl font-extrabold text-gray-800">Історія всіх поїздок 🗺️</h2>
      </div>
      
      <div className="overflow-hidden rounded-xl bg-white shadow-md mb-12">
        <table className="w-full text-left text-sm text-gray-600">
          <thead className="bg-gray-50 text-xs uppercase text-gray-700 border-b border-gray-200">
            <tr>
              <th className="px-6 py-4">ID</th>
              <th className="px-6 py-4">Користувач</th>
              <th className="px-6 py-4">Транспорт</th>
              <th className="px-6 py-4">Статус / Час</th>
              <th className="px-6 py-4 text-right">Сума (₴)</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100">
            {ridesData?.map((ride) => (
              <tr 
                key={ride.id} 
                className="hover:bg-gray-50 transition cursor-pointer" 
                onClick={() => setSelectedRide(ride)}
                title="Клікніть для перегляду маршруту на мапі"
              >
                <td className="px-6 py-4 font-bold text-gray-800">#{ride.id}</td>
                <td className="px-6 py-4">
                  <div className="text-gray-800">{ride.user_email}</div>
                  <div className="text-xs text-gray-500">ID: {ride.user_id}</div>
                </td>
                <td className="px-6 py-4">
                  <div className="font-medium text-gray-800">#{ride.vehicle_id}</div>
                  <div className="text-xs text-gray-400 font-mono">{ride.vehicle_uuid.split('-')[0]}...</div>
                </td>
                <td className="px-6 py-4">
                  <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${ride.status === 'completed' ? 'bg-green-100 text-green-800' : ride.status === 'active' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-800'}`}>
                    {ride.status === 'completed' ? 'Завершена' : ride.status === 'active' ? 'Активна' : ride.status}
                  </span>
                  <div className="mt-1 text-xs text-gray-500">
                    Початок: {new Date(ride.start_time).toLocaleString('uk-UA')}
                  </div>
                  {ride.end_time && (
                    <div className="text-xs text-gray-500">
                      Кінець: {new Date(ride.end_time).toLocaleString('uk-UA')}
                    </div>
                  )}
                </td>
                <td className="px-6 py-4 text-right font-bold text-gray-800">
                  {Number(ride.total_price).toFixed(2)} ₴
                </td>
              </tr>
            ))}
            {(!ridesData || ridesData.length === 0) && (
              <tr><td colSpan={5} className="px-6 py-4 text-center">Немає поїздок у системі</td></tr>
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
                    // Автоматично обираємо перший тариф з нового виду, щоб не відправити помилковий
                    const firstTariff = tariffs?.find(t => t.vehicle_type === newType);
                    setNewVehicleData(prev => ({ 
                      ...prev, 
                      model_id: firstModel ? firstModel.id : prev.model_id,
                      tariff_id: firstTariff ? firstTariff.id : prev.tariff_id
                    }));
                  }}
                  className="w-full cursor-pointer rounded-lg border border-gray-300 bg-white px-3 py-2 outline-none focus:border-gray-800"
                >
                  <option value="scooter">🛴 Самокат</option>
                  <option value="bike">🚲 Велосипед</option>
                  <option value="moped">🛵 Мопед</option>
                  <option value="monowheel">🛞 Моноколесо</option>
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
                  {tariffs?.filter(t => t.vehicle_type === selectedType).map(t => (
                    <option key={t.id} value={t.id}>{t.name} ({t.unlock_price}₴ + {t.minute_price}₴/хв)</option>
                  ))}
                  {tariffs?.filter(t => t.vehicle_type === selectedType).length === 0 && (
                    <option value={0} disabled>Немає тарифів для цього типу</option>
                  )}
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

      {/* Модальне вікно для редагування тарифу */}
      {isEditTariffModalOpen && editingTariff && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-sm rounded-xl bg-white p-6 shadow-xl">
            <h2 className="mb-4 text-xl font-bold text-gray-800">Редагувати тариф 💸</h2>
            <p className="mb-4 text-sm font-medium text-gray-600">{editingTariff.name}</p>
            
            <div className="flex flex-col gap-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Ціна старту (₴)</label>
                <input type="number" step="0.5" min="0" value={editingTariff.unlock_price} onChange={e => setEditingTariff({...editingTariff, unlock_price: parseFloat(e.target.value) || 0})} className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 transition" />
              </div>
              
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Ціна за хвилину (₴)</label>
                <input type="number" step="0.5" min="0" value={editingTariff.minute_price} onChange={e => setEditingTariff({...editingTariff, minute_price: parseFloat(e.target.value) || 0})} className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 transition" />
              </div>
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button 
                onClick={() => { setIsEditTariffModalOpen(false); setEditingTariff(null); }} 
                className="rounded-lg bg-gray-200 px-4 py-2 font-medium text-gray-700 transition hover:bg-gray-300"
              >
                Скасувати
              </button>
              <button 
                onClick={() => updateTariffMutation.mutate(editingTariff)}
                disabled={updateTariffMutation.isPending}
                className="rounded-lg bg-green-600 px-4 py-2 font-medium text-white transition hover:bg-green-700 disabled:opacity-50"
              >
                {updateTariffMutation.isPending ? 'Збереження...' : 'Зберегти'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Модальне вікно для створення нового тарифу */}
      {isAddTariffModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-sm rounded-xl bg-white p-6 shadow-xl">
            <h2 className="mb-4 text-xl font-bold text-gray-800">Створити новий тариф 💸</h2>
            
            <div className="flex flex-col gap-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Назва тарифу</label>
                <input type="text" value={newTariffData.name} onChange={e => setNewTariffData({...newTariffData, name: e.target.value})} placeholder="Наприклад: Вихідний день" className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 transition" />
              </div>
              
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Прив'язка до виду транспорту</label>
                <select 
                  value={newTariffData.vehicle_type} 
                  onChange={e => setNewTariffData({...newTariffData, vehicle_type: e.target.value})} 
                  className="w-full cursor-pointer rounded-lg border border-gray-300 bg-white px-3 py-2 outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 transition"
                >
                  <option value="scooter">🛴 Самокат</option>
                  <option value="bike">🚲 Велосипед</option>
                  <option value="moped">🛵 Мопед</option>
                  <option value="monowheel">🛞 Моноколесо</option>
                </select>
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Ціна старту (₴)</label>
                <input type="number" step="0.5" min="0" value={newTariffData.unlock_price} onChange={e => setNewTariffData({...newTariffData, unlock_price: parseFloat(e.target.value) || 0})} className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 transition" />
              </div>
              
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Ціна за хвилину (₴)</label>
                <input type="number" step="0.5" min="0" value={newTariffData.minute_price} onChange={e => setNewTariffData({...newTariffData, minute_price: parseFloat(e.target.value) || 0})} className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 transition" />
              </div>
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button 
                onClick={() => setIsAddTariffModalOpen(false)} 
                className="rounded-lg bg-gray-200 px-4 py-2 font-medium text-gray-700 transition hover:bg-gray-300"
              >
                Скасувати
              </button>
              <button 
                onClick={() => addTariffMutation.mutate(newTariffData)}
                disabled={addTariffMutation.isPending || !newTariffData.name.trim()}
                className="rounded-lg bg-green-600 px-4 py-2 font-medium text-white transition hover:bg-green-700 disabled:opacity-50"
              >
                {addTariffMutation.isPending ? 'Створення...' : 'Створити'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Модальне вікно для створення нової моделі */}
      {isAddModelModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-sm rounded-xl bg-white p-6 shadow-xl">
            <h2 className="mb-4 text-xl font-bold text-gray-800">Створити нову модель ⚙️</h2>
            
            <div className="flex flex-col gap-4">
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Назва моделі</label>
                <input type="text" value={newModelData.name} onChange={e => setNewModelData({...newModelData, name: e.target.value})} placeholder="Наприклад: Ninebot Max" className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-gray-800 transition" />
              </div>
              
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Вид транспорту</label>
                <select 
                  value={newModelData.vehicle_type} 
                  onChange={e => setNewModelData({...newModelData, vehicle_type: e.target.value})} 
                  className="w-full cursor-pointer rounded-lg border border-gray-300 bg-white px-3 py-2 outline-none focus:border-gray-800 transition"
                >
                  <option value="scooter">🛴 Самокат</option>
                  <option value="bike">🚲 Велосипед</option>
                  <option value="moped">🛵 Мопед</option>
                  <option value="monowheel">🛞 Моноколесо</option>
                </select>
              </div>

              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Ємність батареї (Вт·год)</label>
                <input type="number" min="0" value={newModelData.battery_capacity_wh} onChange={e => setNewModelData({...newModelData, battery_capacity_wh: parseInt(e.target.value) || 0})} className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-gray-800 transition" />
              </div>
              
              <div>
                <label className="mb-1 block text-sm font-medium text-gray-700">Макс. швидкість (км/год)</label>
                <input type="number" min="0" value={newModelData.max_speed} onChange={e => setNewModelData({...newModelData, max_speed: parseInt(e.target.value) || 0})} className="w-full rounded-lg border border-gray-300 px-3 py-2 outline-none focus:border-gray-800 transition" />
              </div>
            </div>

            <div className="mt-6 flex justify-end gap-3">
              <button onClick={() => setIsAddModelModalOpen(false)} className="rounded-lg bg-gray-200 px-4 py-2 font-medium text-gray-700 transition hover:bg-gray-300">
                Скасувати
              </button>
              <button onClick={() => addModelMutation.mutate(newModelData)} disabled={addModelMutation.isPending || !newModelData.name.trim()} className="rounded-lg bg-gray-800 px-4 py-2 font-medium text-white transition hover:bg-gray-700 disabled:opacity-50">
                {addModelMutation.isPending ? 'Створення...' : 'Створити'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Модальне вікно для перегляду маршруту поїздки */}
      {selectedRide && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 backdrop-blur-sm">
          <div className="w-full max-w-4xl rounded-xl bg-white p-6 shadow-xl flex flex-col h-[90vh]">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-bold text-gray-800">Маршрут поїздки #{selectedRide.id} 🗺️</h2>
              <button onClick={() => setSelectedRide(null)} className="text-gray-500 hover:text-gray-800 text-2xl font-bold transition">✕</button>
            </div>
            
            <div className="flex flex-col sm:flex-row gap-4 mb-4 text-sm bg-gray-50 p-4 rounded-lg border border-gray-100">
              <div className="flex-1">
                <span className="text-gray-500 block mb-1">Користувач</span>
                <span className="font-bold text-gray-800">{selectedRide.user_email}</span>
              </div>
              <div className="flex-1">
                <span className="text-gray-500 block mb-1">Транспорт</span>
                <span className="font-bold text-gray-800">#{selectedRide.vehicle_id} ({selectedRide.vehicle_uuid.split('-')[0]}...)</span>
              </div>
              <div className="flex-1">
                <span className="text-gray-500 block mb-1">Статус</span>
                <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${selectedRide.status === 'completed' ? 'bg-green-100 text-green-800' : selectedRide.status === 'active' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-800'}`}>
                  {selectedRide.status === 'completed' ? 'Завершена' : selectedRide.status === 'active' ? 'Активна' : selectedRide.status}
                </span>
              </div>
              <div className="flex-1">
                <span className="text-gray-500 block mb-1">Сума</span>
                <span className="font-bold text-gray-800">{Number(selectedRide.total_price).toFixed(2)} ₴</span>
              </div>
            </div>

            <div className="relative flex-1 mb-4 rounded-lg overflow-hidden border border-gray-300">
              {selectedRide.track && selectedRide.track.length > 0 ? (
                <MapContainer 
                  center={[selectedRide.track[0].latitude, selectedRide.track[0].longitude]} 
                  zoom={15} 
                  style={{ height: '100%', width: '100%' }}
                >
                  <TileLayer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
                  <Polyline 
                    positions={selectedRide.track.map(t => [t.latitude, t.longitude])} 
                    color="#16a34a" 
                    weight={4}
                    dashArray="10, 10"
                  />
                </MapContainer>
              ) : (
                <div className="flex h-full flex-col items-center justify-center bg-gray-50 text-gray-500">
                  <span className="text-4xl mb-2">📡</span>
                  <span className="font-medium">Немає GPS-даних (телеметрії) для цієї поїздки</span>
                </div>
              )}
            </div>

            <div className="flex justify-end gap-3 shrink-0">
              <button onClick={() => setSelectedRide(null)} className="rounded-lg bg-gray-800 px-6 py-2 font-bold text-white transition hover:bg-gray-700">
                Закрити
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}