-- Додаємо прив'язку до конкретного користувача (NULL означає, що код публічний)
ALTER TABLE promo_codes ADD COLUMN user_id INT REFERENCES users(id) ON DELETE CASCADE DEFAULT NULL;

-- ==========================================
-- ТЕСТОВИЙ ПЕРСОНАЛЬНИЙ ПРОМОКОД
-- ==========================================
-- Даємо Демо Користувачу (id = 2) подарунок на 500 грн, який може використати лише він
INSERT INTO promo_codes (code, type, reward_amount, discount_percent, max_uses, user_id) 
VALUES ('HAPPYBDAY', 'top_up', 500.00, 0, 1, 2);