import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/axios';
import { isAxiosError } from 'axios';

// Описуємо структуру поїздки, яку очікуємо від бекенда
interface Ride {
  id: number;
  vehicle_id: number;
  status: string;
  start_time: string;
  end_time?: string;
  total_price: number;
  end_photo_url?: string;
}

// Описуємо структуру платежу, яку очікуємо від бекенда
interface Payment {
  id: number;
  amount: number;
  type: 'top_up' | 'charge';
  status: string;
  created_at: string;
  external_transaction_id?: string;
}

export default function Profile() {
  const queryClient = useQueryClient();
  const [topUpAmount, setTopUpAmount] = useState(100);

  // Стан для перемикання вкладок
  const [activeTab, setActiveTab] = useState<'rides' | 'payments'>('rides');

  // Отримуємо дані користувача (профіль + баланс)
  const { data: user, isLoading: isUserLoading } = useQuery({
    queryKey: ['profile'],
    queryFn: async () => {
      const res = await api.get('/users/me');
      return res.data;
    }
  });

  // Отримуємо історію поїздок
  const { data: historyList, isLoading: isHistoryLoading } = useQuery<Ride[]>({
    queryKey: ['rides', 'history'],
    queryFn: async () => {
      const res = await api.get('/rides/history');
      // Безпечно витягуємо масив поїздок з відповіді бекенда
      return res.data?.rides || res.data || [];
    }
  });

  // Отримуємо історію транзакцій (платежів)
  const { data: paymentsList, isLoading: isPaymentsLoading } = useQuery<Payment[]>({
    queryKey: ['payments', 'history'],
    queryFn: async () => {
      const res = await api.get('/users/payments');
      return res.data?.payments || res.data || [];
    }
  });

  // Мутація для поповнення балансу
  const topUpMutation = useMutation({
    mutationFn: async (amount: number) => {
      const res = await api.post('/users/topup', { amount });
      return res.data;
    },
    onSuccess: () => {
      alert('Баланс успішно поповнено! 💸');
      queryClient.invalidateQueries({ queryKey: ['profile'] }); // Оновлюємо баланс на екрані
      setTopUpAmount(100); // Скидаємо інпут до стандартної суми
    },
    onError: (error) => {
      const message = isAxiosError(error) ? (error.response?.data as { error?: string })?.error : 'Невідома помилка';
      alert(`Помилка поповнення:\n${message}`);
    }
  });

  if (isUserLoading) return <div className="p-8 text-center font-medium">Завантаження профілю...</div>;

  return (
    <div className="mx-auto max-w-4xl p-6">
      <h1 className="mb-6 text-3xl font-bold text-gray-800">Мій профіль 👤</h1>

      {/* Картка балансу */}
      <div className="mb-8 flex flex-col justify-between items-center rounded-xl bg-white p-6 shadow-md border-l-4 border-green-500 sm:flex-row gap-4">
        <div>
          <h2 className="text-lg text-gray-600 font-medium">Поточний баланс:</h2>
          <p className={`text-4xl font-bold ${user?.balance < 0 ? 'text-red-500' : 'text-green-600'}`}>
            {user?.balance?.toFixed(2)} ₴
          </p>
          {user?.balance < 0 && <p className="text-sm text-red-500 mt-1">Будь ласка, погасіть заборгованість!</p>}
        </div>
        
        <div className="flex flex-col sm:flex-row gap-2">
          <input 
            type="number" 
            value={topUpAmount}
            onChange={(e) => setTopUpAmount(Number(e.target.value))}
            className="rounded border-2 border-gray-200 p-2 text-center w-32 focus:border-green-500 outline-none"
            min="10"
            step="10"
          />
          <button 
            onClick={() => topUpMutation.mutate(topUpAmount)}
            disabled={topUpMutation.isPending || topUpAmount <= 0}
            className="rounded bg-green-500 px-6 py-2 text-white font-medium shadow-sm hover:bg-green-600 transition disabled:opacity-50"
          >
            {topUpMutation.isPending ? 'Обробка...' : 'Поповнити'}
          </button>
        </div>
      </div>

      {/* Навігація по вкладках */}
      <div className="mb-6 flex gap-6 border-b border-gray-200">
        <button 
          onClick={() => setActiveTab('rides')}
          className={`pb-3 text-xl font-bold transition ${activeTab === 'rides' ? 'border-b-4 border-green-500 text-gray-800' : 'text-gray-400 hover:text-gray-600'}`}
        >
          Історія поїздок 🛴
        </button>
        <button 
          onClick={() => setActiveTab('payments')}
          className={`pb-3 text-xl font-bold transition ${activeTab === 'payments' ? 'border-b-4 border-green-500 text-gray-800' : 'text-gray-400 hover:text-gray-600'}`}
        >
          Транзакції 💳
        </button>
      </div>

      {/* Контент вкладок */}
      <div className="rounded-xl bg-white shadow-md overflow-hidden">
        
        {/* Вкладка 1: Поїздки */}
        {activeTab === 'rides' && (
          isHistoryLoading ? (
            <p className="p-6 text-center text-gray-500">Завантаження історії...</p>
          ) : !historyList || historyList.length === 0 ? (
            <p className="p-6 text-center text-gray-500">У вас ще немає завершених поїздок.</p>
          ) : (
            <div className="divide-y divide-gray-100">
              {historyList.map((ride) => (
                <div key={ride.id} className="flex flex-col sm:flex-row items-start sm:items-center justify-between p-5 hover:bg-gray-50 transition">
                  <div className="flex items-center gap-4 mb-3 sm:mb-0">
                    {/* Мініатюра фото паркування або заглушка */}
                    {ride.end_photo_url ? (
                      <img 
                        src={`http://localhost:8080${ride.end_photo_url}`} 
                        alt="Паркування" 
                        className="h-16 w-16 rounded-md object-cover border border-gray-200 shadow-sm flex-shrink-0"
                      />
                    ) : (
                      <div className="h-16 w-16 flex flex-shrink-0 items-center justify-center rounded-md bg-gray-100 border border-gray-200 text-2xl shadow-sm">
                        🛴
                      </div>
                    )}
                    
                    <div>
                      <h3 className="font-bold text-gray-800">Поїздка #{ride.id}</h3>
                      <p className="text-sm text-gray-500">
                        {new Date(ride.start_time).toLocaleString('uk-UA', { 
                          day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit' 
                        })}
                      </p>
                      <span className={`inline-block mt-1 px-2 py-0.5 text-xs font-semibold rounded-full ${
                        ride.status === 'completed' ? 'bg-green-100 text-green-700' :
                        ride.status === 'active' ? 'bg-blue-100 text-blue-700' :
                        'bg-gray-100 text-gray-700'
                      }`}>
                        {ride.status === 'completed' ? 'Завершена' : ride.status === 'active' ? 'Активна' : ride.status}
                      </span>
                    </div>
                  </div>
                  
                  <div className="text-right w-full sm:w-auto">
                    <p className="text-xl font-bold text-gray-800">{ride.total_price} ₴</p>
                  </div>
                </div>
              ))}
            </div>
          )
        )}

        {/* Вкладка 2: Транзакції */}
        {activeTab === 'payments' && (
          isPaymentsLoading ? (
            <p className="p-6 text-center text-gray-500">Завантаження транзакцій...</p>
          ) : !paymentsList || paymentsList.length === 0 ? (
            <p className="p-6 text-center text-gray-500">У вас ще немає історії платежів.</p>
          ) : (
            <div className="divide-y divide-gray-100">
              {paymentsList.map((payment) => (
                <div key={payment.id} className="flex items-center justify-between p-5 hover:bg-gray-50 transition">
                  <div className="flex items-center gap-4">
                    <div className={`flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full ${payment.type === 'top_up' ? 'bg-green-100 text-green-600' : 'bg-red-100 text-red-600'} text-xl`}>
                      {payment.type === 'top_up' ? '↓' : '↑'}
                    </div>
                    <div>
                      <h3 className="font-bold text-gray-800">
                        {payment.type === 'top_up' ? 'Поповнення балансу' : 'Оплата поїздки'}
                      </h3>
                      <p className="text-sm text-gray-500">
                        {new Date(payment.created_at).toLocaleString('uk-UA', { 
                          day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit' 
                        })}
                      </p>
                      {payment.external_transaction_id && (
                        <p className="text-xs text-gray-400 font-mono mt-0.5">ID: {payment.external_transaction_id}</p>
                      )}
                    </div>
                  </div>
                  <div className="text-right">
                    <p className={`text-xl font-bold ${payment.type === 'top_up' ? 'text-green-600' : 'text-gray-800'}`}>
                      {payment.type === 'top_up' ? '+' : '-'}{payment.amount} ₴
                    </p>
                    <span className="text-xs text-gray-400 uppercase tracking-wider">{payment.status}</span>
                  </div>
                </div>
              ))}
            </div>
          )
        )}
      </div>
    </div>
  );
}