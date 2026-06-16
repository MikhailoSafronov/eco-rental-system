import { useState, useEffect, useMemo } from 'react'
import { MapContainer, TileLayer, Marker, Popup, GeoJSON, Polyline, useMapEvents } from 'react-leaflet'
import L from 'leaflet'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/axios'
import { isAxiosError } from 'axios'
import type { GeoJsonObject } from 'geojson'

// Інтерфейс для транспорту (узгодьте поля з тим, як бекенд повертає дані)
interface Vehicle {
  id: number;
  vehicle_type: 'scooter' | 'bike' | 'moped' | 'monowheel';
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
  geojson: GeoJsonObject; // Строгий тип для географічних даних
}

// Інтерфейс для історії поїздок
interface Ride {
  id: number;
  status: string;
  start_time: string;
  vehicle_uuid?: string;
  current_lat?: number;
  current_lon?: number;
  battery_level?: number;
  vehicle_type?: 'scooter' | 'bike' | 'moped' | 'monowheel';
  model_name?: string;
  track?: { latitude: number; longitude: number; timestamp: string }[];
}

// Функція для створення іконки залежно від типу транспорту
const getVehicleIcon = (type: string, isHighlight = false) => {
  const emojiMap: Record<string, string> = {
    scooter: '🛴',
    bike: '🚲',
    moped: '🛵',
    monowheel: '🛞',
  };
  
  // Якщо тип не знайдено в словнику, ставимо знак питання
  const emoji = emojiMap[type] || '❓';

  const style = isHighlight 
    ? 'font-size: 36px; filter: drop-shadow(0px 0px 10px rgba(59,130,246,0.9)); transition: all 0.3s; transform: scale(1.1);' 
    : 'font-size: 28px; filter: drop-shadow(0px 2px 2px rgba(0,0,0,0.3)); transition: all 0.3s;';

  return new L.DivIcon({
    html: `<div style="${style}">${emoji}</div>`,
    className: 'bg-transparent border-none',
    iconSize: isHighlight ? [40, 40] : [30, 30],
    iconAnchor: isHighlight ? [20, 40] : [15, 30],
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
  
  // Отримуємо профіль користувача для перевірки статусу блокування
  const { data: user } = useQuery<{ is_blocked: boolean }>({
    queryKey: ['profile'],
    queryFn: async () => {
      const res = await api.get('/users/me')
      return res.data
    }
  })

  // Стан для фотографії паркування
  const [selectedPhoto, setSelectedPhoto] = useState<File | null>(null)
  const [isUploading, setIsUploading] = useState(false)

  // Стан для відображення координат при кліку на мапі
  const [clickedCoords, setClickedCoords] = useState<[number, number] | null>(null)

  // Стани для фільтрації та сортування на мапі
  const [filterType, setFilterType] = useState<string>('all')
  const [maxUnlockPrice, setMaxUnlockPrice] = useState<number | ''>('')
  const [maxMinutePrice, setMaxMinutePrice] = useState<number | ''>('')
  const [isFiltersExpanded, setIsFiltersExpanded] = useState(true)
  const [isSimulating, setIsSimulating] = useState(false)

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
      // Вказуємо структуру сирих даних, що приходять з бекенду (geojson може бути рядком)
      const zones: Array<{ id: number; name: string; geojson: string | GeoJsonObject }> = response.data?.zones || response.data || []
      
      // Підстраховка: PostGIS часто віддає geojson у вигляді рядка, а Leaflet вимагає об'єкт
      return zones.map((z) => ({
        ...z,
        geojson: (typeof z.geojson === 'string' ? JSON.parse(z.geojson) : z.geojson) as GeoJsonObject
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

  // Логіка фільтрації та сортування транспорту (виконується ефективно через useMemo)
  const processedVehicles = useMemo(() => {
    if (!vehicles) return [];
    let result = [...vehicles];

    // Фільтрація за типом
    if (filterType !== 'all') {
      result = result.filter((v) => v.vehicle_type === filterType);
    }

    // Фільтрація за максимальною ціною старту (якщо вказана)
    if (maxUnlockPrice !== '') {
      result = result.filter((v) => v.unlock_price <= Number(maxUnlockPrice));
    }

    // Фільтрація за максимальною ціною за хвилину (якщо вказана)
    if (maxMinutePrice !== '') {
      result = result.filter((v) => v.minute_price <= Number(maxMinutePrice));
    }

    return result;
  }, [vehicles, filterType, maxUnlockPrice, maxMinutePrice]);

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
      setIsSimulating(false) // Автоматично вимикаємо режим симуляції
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

  // Мутація для відправки нових координат на бекенд (IoT телеметрія)
  const simulateMovementMutation = useMutation({
    mutationFn: async (data: { lat: number; lon: number; battery: number }) => {
      if (!activeRide?.vehicle_uuid) return;
      await api.patch(`/iot/vehicles/${activeRide.vehicle_uuid}/telemetry`, {
        latitude: data.lat,
        longitude: data.lon,
        battery_level: data.battery,
        battery: data.battery
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rides', 'history'] });
    },
    onError: () => {
      alert('Помилка симуляції телеметрії.');
    }
  });

  return (
    // Використовуємо прямий style замість Tailwind, щоб гарантувати висоту
    <div style={{ height: '85vh', width: '100%', display: 'block', position: 'relative' }}>
      
      {/* Банер про блокування на головній мапі */}
      {user?.is_blocked && (
        <div className="absolute top-20 left-1/2 -translate-x-1/2 z-[2000] w-11/12 max-w-sm rounded-xl border-l-4 border-red-500 bg-white/95 backdrop-blur p-4 shadow-xl">
          <div className="flex items-center gap-3">
            <span className="text-3xl">🚫</span>
            <div>
              <h3 className="text-sm font-bold text-red-800">Ваш акаунт заблоковано</h3>
              <p className="mt-1 text-xs text-red-700">Оренда транспорту тимчасово недоступна.</p>
            </div>
          </div>
        </div>
      )}

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

      {/* Панель фільтрації та сортування */}
      <div className={`absolute top-4 right-4 z-[1000] flex flex-col gap-4 rounded-xl border border-black/10 bg-white/80 p-4 shadow-xl backdrop-blur-sm transition-all duration-300 ${isFiltersExpanded ? 'w-64' : 'w-auto min-w-[140px]'}`}>
        <div 
          className="flex cursor-pointer items-center justify-between gap-4"
          onClick={() => setIsFiltersExpanded(!isFiltersExpanded)}
        >
          <div>
            <h3 className="text-base font-bold text-gray-800">Фільтри</h3>
            {isFiltersExpanded && <p className="text-xs text-gray-500">Знайдіть ідеальний транспорт</p>}
          </div>
          <button className="flex h-8 w-8 items-center justify-center rounded-full hover:bg-gray-200/50 transition">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" className={`w-5 h-5 text-gray-500 transition-transform duration-300 ${isFiltersExpanded ? 'rotate-180' : 'rotate-0'}`}>
              <path fillRule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z" clipRule="evenodd" />
            </svg>
          </button>
        </div>
        
        {isFiltersExpanded && (
          <>
            {/* Фільтр за типом */}
            <div className="flex gap-1 rounded-lg bg-gray-200/70 p-1">
              <button onClick={() => setFilterType('all')} className={`flex-1 rounded-md py-1 text-sm font-bold transition ${filterType === 'all' ? 'bg-white shadow text-green-600' : 'text-gray-600 hover:bg-white/50'}`}>Всі</button>
              <button onClick={() => setFilterType('scooter')} className={`flex-1 rounded-md py-1 text-lg transition ${filterType === 'scooter' ? 'bg-white shadow' : 'opacity-60 hover:opacity-100'}`}>🛴</button>
              <button onClick={() => setFilterType('bike')} className={`flex-1 rounded-md py-1 text-lg transition ${filterType === 'bike' ? 'bg-white shadow' : 'opacity-60 hover:opacity-100'}`}>🚲</button>
              <button onClick={() => setFilterType('moped')} className={`flex-1 rounded-md py-1 text-lg transition ${filterType === 'moped' ? 'bg-white shadow' : 'opacity-60 hover:opacity-100'}`}>🛵</button>
              <button onClick={() => setFilterType('monowheel')} className={`flex-1 rounded-md py-1 text-lg transition ${filterType === 'monowheel' ? 'bg-white shadow' : 'opacity-60 hover:opacity-100'}`}>🛞</button>
            </div>

            {/* Фільтр за ціною */}
            <div className="flex flex-col gap-3">
              <label className="text-xs font-semibold text-gray-600 block">Ціна, не більше ніж:</label>
              <div className="relative">
                <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" className="w-4 h-4 text-gray-400">
                    <path d="M14 6a4 4 0 0 1-4.899 3.899l-4.226 4.225a.75.75 0 0 1-1.061-1.06l4.225-4.226A4 4 0 1 1 14 6Zm-4-2.5a2.5 2.5 0 1 1 0 5 2.5 2.5 0 0 1 0-5Z" />
                  </svg>
                </div>
                <input 
                  type="number" 
                  min="0"
                  value={maxUnlockPrice} 
                  onChange={(e) => setMaxUnlockPrice(e.target.value === '' ? '' : Number(e.target.value))}
                  className="w-full rounded-lg border border-gray-200 bg-white py-2 pl-9 pr-7 text-sm outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 transition-colors"
                  placeholder="Старт"
                />
                <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                  <span className="text-gray-500 sm:text-sm">₴</span>
                </div>
              </div>
              <div className="relative">
                <div className="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3">
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" className="w-4 h-4 text-gray-400">
                    <path fillRule="evenodd" d="M8 15A7 7 0 1 0 8 1a7 7 0 0 0 0 14Zm.75-10.25a.75.75 0 0 0-1.5 0v4.5c0 .414.336.75.75.75h4.5a.75.75 0 0 0 0-1.5h-3.75V4.75Z" clipRule="evenodd" />
                  </svg>
                </div>
                <input 
                  type="number" 
                  min="0"
                  value={maxMinutePrice} 
                  onChange={(e) => setMaxMinutePrice(e.target.value === '' ? '' : Number(e.target.value))}
                  className="w-full rounded-lg border border-gray-200 bg-white py-2 pl-9 pr-7 text-sm outline-none focus:border-green-500 focus:ring-1 focus:ring-green-500 transition-colors"
                  placeholder="Хвилина"
                />
                <div className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3">
                  <span className="text-gray-500 sm:text-sm">₴</span>
                </div>
              </div>
            </div>
          </>
        )}
      </div>

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

        {/* Відмальовуємо лінію маршруту для активної поїздки */}
        {activeRide?.track && activeRide.track.length > 0 && (
          <Polyline 
            positions={activeRide.track.map(t => [t.latitude, t.longitude])} 
            color="#16a34a" 
            weight={3}
            opacity={0.8}
            dashArray="10, 10"
          />
        )}

        {/* Маркер активного транспорту (його можна тягати) */}
        {activeRide && activeRide.current_lat && activeRide.current_lon && (
          <Marker 
            position={[activeRide.current_lat, activeRide.current_lon]}
            draggable={isSimulating}
            zIndexOffset={1000} // Завжди поверх іншого транспорту
            eventHandlers={{
              dragend: (e) => {
                const marker = e.target;
                const position = marker.getLatLng();
                // Рахуємо дистанцію в метрах (вбудована функція Leaflet)
                const oldLatLng = L.latLng(activeRide.current_lat!, activeRide.current_lon!);
                const distance = oldLatLng.distanceTo(position);
                
                // Мінусуємо 1% батареї за кожні 500 метрів
                const batteryDrop = Math.floor(distance / 500);
                const newBattery = Math.max(0, (activeRide.battery_level || 100) - batteryDrop);

                simulateMovementMutation.mutate({ lat: position.lat, lon: position.lng, battery: newBattery });
              }
            }}
            icon={getVehicleIcon(activeRide.vehicle_type || 'scooter', isSimulating)}
          >
            <Popup>
              <div className="text-center">
                <p className="font-bold text-blue-600">Ваш {activeRide.model_name || 'Транспорт'}</p>
                <p className="text-sm font-medium mt-1">🔋 Заряд: {activeRide.battery_level}%</p>
                {isSimulating && <p className="text-xs text-gray-500 mt-2">Перетягніть іконку по мапі 🖱️</p>}
              </div>
            </Popup>
          </Marker>
        )}

        {/* Динамічно відмальовуємо маркери з отриманих даних */}
        {processedVehicles?.map((vehicle) => (
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
                  disabled={startRideMutation.isPending || user?.is_blocked}
                  className={`w-full rounded px-3 py-2 font-medium text-white shadow-sm transition disabled:opacity-50 disabled:cursor-not-allowed ${user?.is_blocked ? 'bg-red-500' : 'bg-green-500 hover:bg-green-600'}`}
                >
                  {user?.is_blocked ? '🚫 Заблоковано' : startRideMutation.isPending ? 'Запуск...' : 'Орендувати зараз'}
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
          
          {/* Блок симуляції IoT телеметрії */}
          <div className="flex flex-col sm:flex-row items-center justify-between rounded-xl bg-blue-50 p-4 border border-blue-100 gap-3">
            <div>
              <p className="text-sm font-bold text-blue-800">Тестування IoT 📡</p>
              <p className="text-xs text-blue-600">Перетягніть транспорт по мапі</p>
            </div>
            <button 
              onClick={() => setIsSimulating(!isSimulating)}
              disabled={simulateMovementMutation.isPending}
              className={`rounded-lg px-4 py-2 text-sm font-bold text-white shadow-sm transition disabled:opacity-50 ${isSimulating ? 'bg-orange-500 hover:bg-orange-600' : 'bg-blue-600 hover:bg-blue-700'}`}
            >
              {isSimulating ? 'Симуляція активна' : 'Симулювати рух'}
            </button>
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