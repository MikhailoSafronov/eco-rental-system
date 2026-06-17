-- Додаємо тип промокоду та відсоток знижки
ALTER TABLE promo_codes ADD COLUMN type VARCHAR(20) NOT NULL DEFAULT 'top_up';
ALTER TABLE promo_codes ADD COLUMN discount_percent INT NOT NULL DEFAULT 0;

-- Додаємо поле користувачу для зберігання активної знижки на наступну поїздку
ALTER TABLE users ADD COLUMN active_discount_percent INT NOT NULL DEFAULT 0;

-- ==========================================
-- ТЕСТОВІ ПРОМОКОДИ
-- ==========================================
INSERT INTO promo_codes (code, type, reward_amount, discount_percent, max_uses) VALUES 
('ECO2026', 'top_up', 150.00, 0, 1000),      -- Дає 150 грн на баланс
('MINUS50', 'discount', 0, 50, 1000),        -- Дає знижку 50% на наступну поїздку
('WELCOME', 'top_up', 50.00, 0, 9999)        -- Дає 50 грн (майже безлімітний)
ON CONFLICT (code) DO NOTHING;               -- Захист від помилок, якщо код вже існує