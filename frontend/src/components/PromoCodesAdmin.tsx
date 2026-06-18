import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/axios';
import { isAxiosError } from 'axios';

interface PromoCode {
  id: number;
  code: string;
  type: 'top_up' | 'discount';
  reward_amount: number;
  discount_percent: number;
  max_uses: number;
  current_uses: number;
  user_id: number | null;
  user_email: string | null;
}

export const PromoCodesAdmin = () => {
  const queryClient = useQueryClient();
  const [newPromo, setNewPromo] = useState({
    code: '',
    type: 'top_up',
    reward_amount: 100,
    discount_percent: 0,
    max_uses: 100,
    user_id: ''
  });

  const { data: promos, isLoading } = useQuery<PromoCode[]>({
    queryKey: ['admin', 'promocodes'],
    queryFn: async () => {
      const res = await api.get('/admin/promocodes');
      return res.data || [];
    }
  });

  const createMutation = useMutation({
    mutationFn: async (promo: typeof newPromo) => {
      const payload = {
        ...promo,
        user_id: promo.user_id ? parseInt(promo.user_id) : null
      };
      await api.post('/admin/promocodes', payload);
    },
    onSuccess: () => {
      alert('🎉 Промокод успішно створено!');
      queryClient.invalidateQueries({ queryKey: ['admin', 'promocodes'] });
      setNewPromo({ code: '', type: 'top_up', reward_amount: 100, discount_percent: 0, max_uses: 100, user_id: '' });
    },
    onError: (err) => {
      const msg = isAxiosError(err) ? err.response?.data?.error : 'Невідома помилка';
      alert(`❌ Помилка: ${msg}`);
    }
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      await api.delete(`/admin/promocodes/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'promocodes'] });
    }
  });

  const handleDelete = (id: number) => {
    if (window.confirm('Ви впевнені, що хочете видалити цей промокод?')) {
      deleteMutation.mutate(id);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    createMutation.mutate(newPromo);
  };

  if (isLoading) return <div className="p-8 text-center text-gray-500">Завантаження промокодів...</div>;

  return (
    <div className="space-y-8">
      {/* Форма створення промокоду */}
      <div className="rounded-xl border border-gray-100 bg-white p-6 shadow-md">
        <h2 className="mb-6 text-xl font-bold text-gray-800">Створити новий промокод 🎁</h2>
        <form onSubmit={handleSubmit} className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <div>
            <label className="mb-1 block text-sm font-semibold text-gray-600">Код (наприклад: SUMMER2026)</label>
            <input required type="text" value={newPromo.code} onChange={e => setNewPromo({...newPromo, code: e.target.value.toUpperCase()})} className="w-full rounded-lg border border-gray-300 p-2 font-mono uppercase focus:border-green-500 focus:outline-none" placeholder="Введіть код..." />
          </div>
          <div>
            <label className="mb-1 block text-sm font-semibold text-gray-600">Тип бонусу</label>
            <select value={newPromo.type} onChange={e => setNewPromo({...newPromo, type: e.target.value as 'top_up' | 'discount'})} className="w-full rounded-lg border border-gray-300 p-2 focus:border-green-500 focus:outline-none">
              <option value="top_up">Поповнення балансу (₴)</option>
              <option value="discount">Знижка на поїздку (%)</option>
            </select>
          </div>

          {newPromo.type === 'top_up' ? (
            <div>
              <label className="mb-1 block text-sm font-semibold text-gray-600">Сума поповнення (₴)</label>
              <input required type="number" min="1" value={newPromo.reward_amount} onChange={e => setNewPromo({...newPromo, reward_amount: parseFloat(e.target.value)})} className="w-full rounded-lg border border-gray-300 p-2 focus:border-green-500 focus:outline-none" />
            </div>
          ) : (
            <div>
              <label className="mb-1 block text-sm font-semibold text-gray-600">Відсоток знижки (%)</label>
              <input required type="number" min="1" max="100" value={newPromo.discount_percent} onChange={e => setNewPromo({...newPromo, discount_percent: parseInt(e.target.value)})} className="w-full rounded-lg border border-gray-300 p-2 focus:border-green-500 focus:outline-none" />
            </div>
          )}

          <div>
            <label className="mb-1 block text-sm font-semibold text-gray-600">Ліміт використань (для всіх)</label>
            <input required type="number" min="1" value={newPromo.max_uses} onChange={e => setNewPromo({...newPromo, max_uses: parseInt(e.target.value)})} className="w-full rounded-lg border border-gray-300 p-2 focus:border-green-500 focus:outline-none" />
          </div>

          <div>
            <label className="mb-1 block text-sm font-semibold text-gray-600">ID Користувача (необов'язково)</label>
            <input type="number" min="1" value={newPromo.user_id} onChange={e => setNewPromo({...newPromo, user_id: e.target.value})} className="w-full rounded-lg border border-gray-300 p-2 focus:border-green-500 focus:outline-none" placeholder="Залиште пустим для всіх" />
            <p className="mt-1 text-xs text-gray-500">Вкажіть ID, щоб зробити код персональним</p>
          </div>

          <div className="flex items-end lg:col-span-3">
            <button type="submit" disabled={createMutation.isPending} className="w-full rounded-lg bg-green-600 p-3 font-bold text-white transition hover:bg-green-700 disabled:opacity-50">
              {createMutation.isPending ? 'Створення...' : 'Створити промокод'}
            </button>
          </div>
        </form>
      </div>

      {/* Таблиця промокодів */}
      <div className="overflow-x-auto rounded-xl border border-gray-100 bg-white p-6 shadow-md">
        <h2 className="mb-4 text-xl font-bold text-gray-800">Список промокодів</h2>
        <table className="w-full border-collapse text-left text-sm">
          <thead>
            <tr className="border-b bg-gray-50 text-gray-600">
              <th className="p-3">ID / Код</th>
              <th className="p-3">Бонус</th>
              <th className="p-3">Використано</th>
              <th className="p-3">Прив'язка (Персональні)</th>
              <th className="p-3 text-right">Дії</th>
            </tr>
          </thead>
          <tbody>
            {promos?.map(p => (
              <tr key={p.id} className="border-b transition hover:bg-gray-50">
                <td className="p-3"><span className="text-gray-400">#{p.id}</span> <span className="ml-2 font-mono font-bold text-gray-800">{p.code}</span></td>
                <td className="p-3 font-bold">{p.type === 'top_up' ? <span className="text-green-600">+{p.reward_amount} ₴</span> : <span className="text-blue-600">-{p.discount_percent}% знижки</span>}</td>
                <td className="p-3"><span className={p.current_uses >= p.max_uses ? 'font-bold text-red-500' : 'text-gray-700'}>{p.current_uses} / {p.max_uses}</span></td>
                <td className="p-3">{p.user_id ? <span className="rounded-md bg-purple-100 px-2 py-1 text-xs font-semibold text-purple-700" title={p.user_email || ''}>ID: {p.user_id} ({p.user_email})</span> : <span className="text-gray-400">Загальний</span>}</td>
                <td className="p-3 text-right"><button onClick={() => handleDelete(p.id)} className="font-medium text-red-500 hover:text-red-700">Видалити</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};