import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import './index.css'
import App from './App.tsx'
import 'leaflet/dist/leaflet.css' // Обов'язково імпортуємо стилі для Leaflet

// Створюємо клієнт для React Query з базовими налаштуваннями
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false, // не робити запит заново, коли юзер просто перемикає вкладки браузера
      retry: 1, // при помилці мережі спробувати ще 1 раз
    },
  },
})

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <App />
      </BrowserRouter>
    </QueryClientProvider>
  </StrictMode>,
)
