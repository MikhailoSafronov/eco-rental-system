import { useState, useEffect } from 'react'
import { MapContainer, TileLayer, Marker, Popup, GeoJSON, useMapEvents } from 'react-leaflet'
import L from 'leaflet'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/axios'
import { isAxiosError } from 'axios'

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

// Інтерфейс для паркувальної зони
interface ParkingZone {
  id: number;
  name: string;
  geojson: string | object; // Дані у форматі GeoJSON
}

// Інтерфейс для історії поїздок
interface Ride {
  id: number;
  status: string;
  start_time: string;
}

// Функція для створення іконки залежно від типу транспорту
const getVehicleIcon = (type: string) => {
  const emojiMap: Record<string, string> = {
    scooter: '🛴',
    bike: '🚲',
    moped: '🛵',
  };
  
  // Якщо тип не знайдено в словнику, ставимо знак питання
  const emoji = emojiMap[type] || '❓';

  return new L.DivIcon({
    html: `<div style="font-size: 28px; filter: drop-shadow(0px 2px 2px rgba(0,0,0,0.3));">${emoji}</div>`,
    className: 'bg-transparent border-none',
    iconSize: [30, 30],
    iconAnchor: [15, 30],
  });
};

// Допоміжний компонент для перехоплення кліків по мапі
function MapClickHandler({ onMapClick }: { onMapClick: (lat: number, lng: number) => void }) {
  useMapEvents({
    click(e) {
      onMapClick(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

export default function Home() {
  // Отримуємо доступ до кешу React Query
  const queryClient = useQueryClient()
  
  // Стан для фотографії паркування
  const [selectedPhoto, setSelectedPhoto] = useState<File | null>(null)
  const [isUploading, setIsUploading] = useState(false)

  // Стан для відображення координат при кліку на мапі
  const [clickedCoords, setClickedCoords] = useState<[number, number] | null>(null)

  // Автоматично приховуємо цифри координат через 2.5 секунди
  useEffect(() => {
    if (clickedCoords) {
      const timer = setTimeout(() => setClickedCoords(null), 2500)
      return () => clearTimeout(timer)
    }
  }, [clickedCoords])

  // Координати центру Херсона (беремо з ваших тестових даних у БД)
  const centerPosition: [number, number] = [46.6322, 32.6146]

  // Запит для отримання паркувальних зон
  const { data: parkingZones } = useQuery<ParkingZone[]>({
    queryKey: ['zones'],
    queryFn: async () => {
      const response = await api.get('/zones')
      const zones = response.data?.zones || response.data || []
      
      // Підстраховка: PostGIS часто віддає geojson у вигляді рядка, а Leaflet вимагає об'єкт
      return zones.map((z: ParkingZone) => ({
        ...z,
        geojson: typeof z.geojson === 'string' ? JSON.parse(z.geojson) : z.geojson
      }))
    },
    refetchOnWindowFocus: false, // Зони змінюються дуже рідко, тому зайвий раз не оновлюємо
  })

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

  // Запит історії поїздок для пошуку активної
  const { data: historyData } = useQuery<Ride[]>({
    queryKey: ['rides', 'history'],
    queryFn: async () => {
      const response = await api.get('/rides/history')
      
      // Підстраховка: якщо бекенд загортає масив у об'єкт (наприклад, { rides: [...] })
      const data = response.data?.rides || response.data || []
      return Array.isArray(data) ? data : []
    },
    refetchInterval: 10000, // Оновлювати раз на 10 секунд
  })

  // Шукаємо, чи є серед поїздок активна
  const activeRide = historyData?.find((r) => r.status === 'active') || null

  // Мутація для початку оренди
  const startRideMutation = useMutation({
    mutationFn: async (vehicleId: number) => {
      // Відправляємо POST запит на бекенд з ID транспорту
      const response = await api.post('/rides/start', { vehicle_id: vehicleId })
      return response.data
    },
    onSuccess: () => {
      alert('Поїздку успішно розпочато! Гарної дороги 🚀')
      
      // Наказуємо React Query миттєво оновити список транспорту на мапі
      // Орендований транспорт зникне, бо його статус зміниться на 'rented'
      queryClient.invalidateQueries({ queryKey: ['vehicles', 'available'] })
      // ТАКОЖ миттєво оновлюємо історію, щоб з'явилася картка "Активна поїздка"
      queryClient.invalidateQueries({ queryKey: ['rides', 'history'] })
    },
    onError: (error: Error) => {
      // Безпечно витягуємо повідомлення про помилку від нашого Go бекенду
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Невідома помилка при спробі оренди'
      alert(`Не вдалося почати оренду:\n${message}`)
    }
  })

  // Мутація для завершення оренди
  const endRideMutation = useMutation({
    mutationFn: async (photoUrl: string) => {
      // Відправляємо POST запит на бекенд із завантаженим фото
      // ВИПРАВЛЕНО: ключ тепер end_photo_url, як очікує бекенд
      const response = await api.post('/rides/end', { end_photo_url: photoUrl })
      return response.data
    },
    onSuccess: () => {
      alert('Поїздку успішно завершено! Чек сформовано 🎉')
      setSelectedPhoto(null) // Очищуємо фото
      queryClient.invalidateQueries({ queryKey: ['rides', 'history'] })
      queryClient.invalidateQueries({ queryKey: ['vehicles', 'available'] })
    },
    onError: (error: Error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Невідома помилка'
      alert(`Не вдалося завершити оренду:\n${message}`)
    }
  })

  // Функція обробки завершення (завантажуємо фото, потім завершуємо поїздку)
  const handleEndRide = async () => {
    if (!selectedPhoto) {
      alert('Будь ласка, зробіть фотографію припаркованого транспорту перед завершенням поїздки 📸')
      return
    }

    try {
      setIsUploading(true)
      const formData = new FormData()
      formData.append('photo', selectedPhoto) // Припускаємо, що бекенд чекає поле 'photo'
      
      const uploadRes = await api.post('/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })

      // ВИПРАВЛЕНО: читаємо лише поле 'url', яке повертає наш upload.go
      const photoUrl = (uploadRes.data as { url: string }).url
      endRideMutation.mutate(photoUrl)
    } catch (error) {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Помилка мережі'
      alert('Не вдалося завантажити фото:\n' + message)
    } finally {
      setIsUploading(false)
    }
  }

  return (
    // Використовуємо прямий style замість Tailwind, щоб гарантувати висоту
    <div style={{ height: '85vh', width: '100%', display: 'block', position: 'relative' }}>
      
      {/* Індикатори стану */}
      {isLoading && (
        <div className="absolute top-4 left-1/2 -translate-x-1/2 z-50 rounded-md bg-white px-4 py-2 shadow-md font-medium text-green-700">
          Шукаємо транспорт поруч... 🛴🚲🛵
        </div>
      )}
      {isError && (
        <div className="absolute top-4 left-1/2 -translate-x-1/2 z-50 rounded-md bg-red-100 px-4 py-2 font-medium text-red-600 shadow-md">
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

        {/* Обробник кліків для копіювання координат */}
        <MapClickHandler onMapClick={(lat, lng) => {
          const coords = `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
          navigator.clipboard.writeText(coords); // Копіюємо у буфер обміну
          setClickedCoords([lat, lng]); // Зберігаємо координати для показу на мапі
        }} />

        {/* Показуємо координати прямо на мапі без вікна */}
        {clickedCoords && (
          <Marker 
            position={clickedCoords}
            interactive={false} // Робимо напис "прозорим" для кліків, щоб не блокував мапу
            icon={new L.DivIcon({
              html: `<div class="font-mono font-bold text-green-700 text-sm whitespace-nowrap">${clickedCoords[0].toFixed(5)}, ${clickedCoords[1].toFixed(5)}</div>`,
              className: 'bg-transparent border-none',
              iconAnchor: [55, 10], // Приблизно центруємо напис відносно точки кліку
            })}
          />
        )}

        {/* Відмальовуємо зелені паркувальні зони */}
        {parkingZones?.map((zone) => (
          <GeoJSON
            key={`zone-${zone.id}`}
            data={zone.geojson}
            style={{
              color: '#22c55e', // Зелений колір межі (Tailwind green-500)
              weight: 2,        // Товщина лінії
              opacity: 0.8,     // Прозорість лінії
              fillColor: '#22c55e', // Колір заливки
              fillOpacity: 0.2, // Прозорість заливки
            }}
          >
            <Popup>{zone.name}</Popup>
          </GeoJSON>
        ))}

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

                <button 
                  onClick={() => startRideMutation.mutate(vehicle.id)}
                  disabled={startRideMutation.isPending}
                  className="w-full rounded bg-green-500 px-3 py-2 font-medium text-white shadow-sm transition hover:bg-green-600 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {startRideMutation.isPending ? 'Запуск...' : 'Орендувати зараз'}
                </button>
              </div>
            </Popup>
          </Marker>
        ))}
      </MapContainer>

      {/* Банер активної поїздки (Floating Card) */}
      {activeRide && (
        <div className="absolute bottom-8 right-8 z-50 w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl border-t-[6px] border-green-500 flex flex-col gap-5">
          <div className="text-center border-b border-gray-200 pb-4">
            <h2 className="text-3xl font-extrabold text-gray-800">🛴 Активна поїздка</h2>
            <p className="mt-2 text-base font-medium text-gray-500">
              Час старту: {new Date(activeRide.start_time).toLocaleTimeString()}
            </p>
          </div>
          
          {/* Блок завантаження фото */}
          <div className="flex w-full flex-col">
            {!selectedPhoto ? (
              <label className="flex w-full cursor-pointer items-center justify-center rounded-xl border-2 border-dashed border-gray-400 bg-gray-50 px-6 py-8 shadow-sm transition hover:bg-gray-100">
                <span className="text-xl font-bold text-gray-700">📸 Додати фото паркування</span>
                <input 
                  type="file" 
                  accept="image/*" 
                  capture="environment" 
                  style={{ display: 'none' }} 
                  onChange={(e) => setSelectedPhoto(e.target.files?.[0] || null)} 
                />
              </label>
            ) : (
              <div className="flex w-full items-center gap-4 rounded-xl border border-gray-200 bg-gray-50 p-4 shadow-sm">
                <img 
                  src={URL.createObjectURL(selectedPhoto)} 
                  alt="Preview" 
                  className="flex-shrink-0 rounded-lg border border-gray-300 shadow-sm"
                  style={{ width: '80px', height: '80px', minWidth: '80px', minHeight: '80px', objectFit: 'cover' }}
                />
                <label className="flex flex-1 cursor-pointer items-center justify-center rounded-lg border border-gray-300 bg-white py-4 text-lg font-bold text-gray-700 shadow-sm transition hover:bg-gray-100">
                  🔄 Змінити фото
                  <input 
                    type="file" 
                    accept="image/*" 
                    capture="environment" 
                    style={{ display: 'none' }} 
                    onChange={(e) => setSelectedPhoto(e.target.files?.[0] || null)} 
                  />
                </label>
              </div>
            )}
          </div>

          <button 
            onClick={handleEndRide}
            disabled={endRideMutation.isPending || isUploading}
            className="w-full rounded-xl bg-red-500 py-4 text-xl font-bold text-white shadow-lg transition hover:bg-red-600 active:bg-red-700 disabled:opacity-50"
          >
            {endRideMutation.isPending || isUploading ? 'Обробка...' : 'Завершити поїздку 🛑'}
          </button>
        </div>
      )}
    </div>
  )
}