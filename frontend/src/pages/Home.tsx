import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet'
import L from 'leaflet'
import { useQuery } from '@tanstack/react-query'
import { api } from '../api/axios'

// Інтерфейс для транспорту (узгодьте поля з тим, як бекенд повертає дані)
interface Vehicle {
  id: number;
  vehicle_type: 'scooter' | 'bike' | 'moped';
  model_name?: string; 
  battery_level: number;
  latitude: number;
  longitude: number;
  unlock_price: number;
  minute_price: number;
}

// Функція для створення іконки залежно від типу транспорту
const getVehicleIcon = (type: string) => {
  let emoji = '🛴'; // За замовчуванням самокат
  if (type === 'bike') emoji = '🚲';
  if (type === 'moped') emoji = '🛵';

  return new L.DivIcon({
    html: `<div style="font-size: 28px; filter: drop-shadow(0px 2px 2px rgba(0,0,0,0.3));">${emoji}</div>`,
    className: 'bg-transparent border-none',
    iconSize: [30, 30],
    iconAnchor: [15, 30],
  });
};

export default function Home() {
  // Координати центру Херсона (беремо з ваших тестових даних у БД)
  const centerPosition: [number, number] = [46.6322, 32.6146]

  // Запит для отримання доступного транспорту через React Query
  const { data: vehicles, isLoading, isError } = useQuery({
    queryKey: ['vehicles', 'available'],
    queryFn: async () => {
      // Шлях виправлено відповідно до вашого server.go
      const response = await api.get<Vehicle[]>('/vehicles')
      return response.data
    },
    refetchInterval: 30000, // Автоматично оновлювати мапу кожні 30 секунд
  })

  return (
    // Використовуємо прямий style замість Tailwind, щоб гарантувати висоту
    <div style={{ height: '85vh', width: '100%', display: 'block', position: 'relative' }}>
      
      {/* Індикатори стану */}
      {isLoading && (
        <div className="absolute top-4 left-1/2 -translate-x-1/2 z-[1000] rounded-md bg-white px-4 py-2 shadow-md font-medium text-green-700">
          Шукаємо транспорт поруч... 🛴🚲🛵
        </div>
      )}
      {isError && (
        <div className="absolute top-4 left-1/2 -translate-x-1/2 z-[1000] rounded-md bg-red-100 px-4 py-2 font-medium text-red-600 shadow-md">
          Помилка завантаження мапи транспорту
        </div>
      )}

      <MapContainer 
        center={centerPosition} 
        zoom={14} 
        style={{ height: '100%', width: '100%', zIndex: 0 }}
      >
        <TileLayer
          attribution='&copy; OpenStreetMap'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />

        {/* Динамічно відмальовуємо маркери з отриманих даних */}
        {vehicles?.map((vehicle) => (
          <Marker 
            key={vehicle.id} 
            position={[vehicle.latitude, vehicle.longitude]} 
            icon={getVehicleIcon(vehicle.vehicle_type)}
          >
            <Popup>
              <div className="w-48 text-center">
                <h3 className="mb-2 border-b pb-1 text-lg font-bold text-green-600">
                  {vehicle.model_name || `Транспорт #${vehicle.id}`}
                </h3>
                
                <div className="mb-3 flex flex-col gap-1 text-sm text-gray-700 text-left">
                  <div className="flex justify-between">
                    <span>🔋 Заряд:</span>
                    <span className={`font-semibold ${vehicle.battery_level < 20 ? 'text-red-500' : 'text-green-600'}`}>{vehicle.battery_level}%</span>
                  </div>
                  <div className="flex justify-between">
                    <span>🔓 Старт:</span>
                    <span className="font-semibold">{vehicle.unlock_price} ₴</span>
                  </div>
                  <div className="flex justify-between">
                    <span>⏱ Хвилина:</span>
                    <span className="font-semibold">{vehicle.minute_price} ₴</span>
                  </div>
                </div>

                <button className="w-full rounded bg-green-500 px-3 py-2 font-medium text-white shadow-sm transition hover:bg-green-600">
                  Орендувати зараз
                </button>
              </div>
            </Popup>
          </Marker>
        ))}
      </MapContainer>
    </div>
  )
}