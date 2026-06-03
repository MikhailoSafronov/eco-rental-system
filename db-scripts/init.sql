-- ==========================================
-- 1. НАЛАШТУВАННЯ ТА РОЗШИРЕННЯ
-- ==========================================
-- Підключення розширення PostGIS для роботи з геоданними
CREATE EXTENSION IF NOT EXISTS postgis;

-- ==========================================
-- 2. СТВОРЕННЯ ПЕРЕЛІЧУВАНИХ ТИПІВ (ENUM)
-- ==========================================
CREATE TYPE user_role AS ENUM ('client', 'mechanic', 'admin');
CREATE TYPE vehicle_type AS ENUM ('scooter', 'bike');
CREATE TYPE vehicle_status AS ENUM ('available', 'rented', 'low_battery', 'broken', 'maintenance');
CREATE TYPE ride_status AS ENUM ('active', 'completed', 'cancelled');
CREATE TYPE payment_type AS ENUM ('charge', 'top_up');
CREATE TYPE payment_status AS ENUM ('pending', 'succeeded', 'failed');

-- ==========================================
-- 3. СТВОРЕННЯ ТАБЛИЦЬ
-- ==========================================

-- Довідник користувачів
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE, -- Додано UNIQUE для надійності на рівні таблиці
    phone VARCHAR(20) NOT NULL UNIQUE,  -- Додано UNIQUE для надійності на рівні таблиці
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'client',
    balance NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Довідник зон паркування
CREATE TABLE parking_zones (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    polygon GEOGRAPHY(Polygon, 4326) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Довідник моделей транспорту
CREATE TABLE vehicle_models (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    type vehicle_type NOT NULL,
    battery_capacity_wh INT NOT NULL,
    max_speed INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Довідник тарифів
CREATE TABLE tariffs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    unlock_price NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    minute_price NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ресурс: конкретні транспортні засоби
CREATE TABLE vehicles (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    model_id INT NOT NULL REFERENCES vehicle_models(id) ON DELETE RESTRICT,
    tariff_id INT NOT NULL REFERENCES tariffs(id) ON DELETE RESTRICT,
    location GEOGRAPHY(Point, 4326) NOT NULL,
    battery_level INT NOT NULL DEFAULT 100 CHECK (battery_level >= 0 AND battery_level <= 100),
    status vehicle_status NOT NULL DEFAULT 'available',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Транзакція: поїздки
CREATE TABLE rides (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE RESTRICT,
    status ride_status NOT NULL DEFAULT 'active',
    start_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    end_time TIMESTAMPTZ,
    start_location GEOGRAPHY(Point, 4326) NOT NULL,
    end_location GEOGRAPHY(Point, 4326),
    end_photo_url VARCHAR(255),
    total_price NUMERIC(10, 2) NOT NULL DEFAULT 0.00
);

-- Транзакція: телеметрія (трекінг шляху)
CREATE TABLE ride_telemetry (
    id BIGSERIAL PRIMARY KEY,
    ride_id BIGINT NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    location GEOGRAPHY(Point, 4326) NOT NULL,
    speed INT NOT NULL DEFAULT 0,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Транзакція: платежі
CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    ride_id BIGINT REFERENCES rides(id) ON DELETE SET NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    amount NUMERIC(10, 2) NOT NULL CHECK (amount > 0),
    type payment_type NOT NULL,
    status payment_status NOT NULL DEFAULT 'pending',
    external_transaction_id VARCHAR(100) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Транзакція: журнал обслуговування (ремонт)
CREATE TABLE maintenance_logs (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id) ON DELETE RESTRICT,
    mechanic_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    issue_description TEXT NOT NULL,
    fix_description TEXT,
    cost NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

-- ==========================================
-- 4. СТВОРЕННЯ ІНДЕКСІВ
-- ==========================================

-- Часткові унікальні індекси (для м'якого видалення)
CREATE UNIQUE INDEX idx_users_email_unique_active ON users(email) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_users_phone_unique_active ON users(phone) WHERE deleted_at IS NULL;

-- Захист від Race Condition
CREATE UNIQUE INDEX idx_one_active_ride_per_user ON rides(user_id) WHERE status = 'active';
CREATE UNIQUE INDEX idx_one_active_ride_per_vehicle ON rides(vehicle_id) WHERE status = 'active';

-- Просторові індекси (GiST) для швидкого пошуку по карті
CREATE INDEX idx_parking_zones_polygon ON parking_zones USING gist(polygon);
CREATE INDEX idx_vehicles_location ON vehicles USING gist(location);
CREATE INDEX idx_ride_telemetry_location ON ride_telemetry USING gist(location);

-- B-Tree індекси для частих вибірок
CREATE INDEX idx_vehicles_status ON vehicles(status);
CREATE INDEX idx_ride_telemetry_ride_id ON ride_telemetry(ride_id);
CREATE INDEX idx_payments_user_id ON payments(user_id);

-- ==========================================
-- 5. ТЕСТОВІ ДАНІ (SEED DATA)
-- ==========================================

-- Додаємо базові тарифи
INSERT INTO tariffs (name, unlock_price, minute_price) VALUES
('Базовий (Будні)', 9.00, 3.00),
('Вихідний день', 15.00, 4.00);

-- Додаємо моделі самокатів
INSERT INTO vehicle_models (name, type, battery_capacity_wh, max_speed) VALUES
('Ninebot Max G30', 'scooter', 551, 25),
('Xiaomi Mi Pro 2', 'scooter', 474, 25);

-- Додаємо тестові самокати на вулиці
INSERT INTO vehicles (model_id, tariff_id, location, battery_level, status) VALUES
-- Самокат 1: Готовий до оренди (Майдан Незалежності)
(1, 1, ST_SetSRID(ST_MakePoint(30.523400, 50.450100), 4326), 98, 'available'),

-- Самокат 2: Готовий до оренди, але розряджається (Золоті Ворота)
(1, 1, ST_SetSRID(ST_MakePoint(30.518000, 50.448000), 4326), 45, 'available'),

-- Самокат 3: На ремонті (м. Хрещатик)
(2, 2, ST_SetSRID(ST_MakePoint(30.522000, 50.447000), 4326), 12, 'maintenance');