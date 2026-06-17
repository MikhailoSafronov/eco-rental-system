-- backend/migrations/000003_add_promo_codes.up.sql

CREATE TABLE promo_codes (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    reward_amount NUMERIC(10, 2) NOT NULL, -- Сума, яка додається на баланс
    max_uses INT NOT NULL DEFAULT 100,     -- Максимальна кількість використань (загальна)
    current_uses INT NOT NULL DEFAULT 0,   -- Скільки разів вже використано
    expires_at TIMESTAMP WITH TIME ZONE,   -- Доки діє (NULL = безстроково)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Таблиця для відстеження того, хто вже використав промокод
CREATE TABLE user_promo_usages (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    promo_id INT NOT NULL REFERENCES promo_codes(id) ON DELETE CASCADE,
    used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, promo_id) -- Гарантує, що 1 юзер використає 1 код лише 1 раз
);

-- Тестовий промокод
INSERT INTO promo_codes (code, reward_amount, max_uses) 
VALUES ('ECO2026', 100.00, 500);
