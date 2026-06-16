-- ==========================================
-- ТЕСТОВІ ДАНІ (SEED DATA)
-- ==========================================

-- 1. Додаємо базові тарифи
INSERT INTO tariffs (name, type, unlock_price, minute_price) VALUES
('Базовий (Будні)', 'scooter', 9.00, 3.00),
('Вихідний день', 'scooter', 15.00, 4.00),
('Студентський', 'scooter', 5.00, 2.00),
('Базовий Вело', 'bike', 5.00, 1.50),
('Спортивний Вело', 'bike', 7.00, 2.50),
('Базовий Мопед', 'moped', 20.00, 5.00),
('Базовий Моноколесо', 'monowheel', 12.00, 3.50),
('Екстрим Моноколесо', 'monowheel', 18.00, 4.50);

-- 2. Додаємо моделі транспорту
INSERT INTO vehicle_models (name, type, battery_capacity_wh, max_speed) VALUES
('Ninebot Max G30', 'scooter', 551, 25),
('Xiaomi Mi Pro 2', 'scooter', 474, 25),
('Dualtron Mini', 'scooter', 730, 45),
('E-Bike City Pro', 'bike', 500, 25),
('Gravel Master', 'bike', 600, 30),
('NIU NQi Sport', 'moped', 1440, 45),
('Inmotion V11', 'monowheel', 1500, 50),
('KingSong 16X', 'monowheel', 1554, 45);

-- 3. Додаємо зони паркування
INSERT INTO parking_zones (name, polygon) VALUES
('Центр (Херсон)', ST_GeomFromText('POLYGON((32.605 46.640, 32.622 46.640, 32.625 46.632, 32.615 46.626, 32.608 46.628, 32.605 46.640))', 4326)),
('Залізничний вокзал (Херсон)', ST_GeomFromText('POLYGON((32.595 46.650, 32.615 46.650, 32.610 46.642, 32.598 46.642, 32.595 46.650))', 4326)),
('ТРЦ Фабрика', ST_GeomFromText('POLYGON((32.6410 46.6755, 32.6475 46.6755, 32.6475 46.6705, 32.6410 46.6705, 32.6410 46.6755))', 4326)),
('Таврійський мікрорайон', ST_GeomFromText('POLYGON((32.610 46.670, 32.630 46.670, 32.630 46.660, 32.610 46.660, 32.610 46.670))', 4326)),
('Шуменський парк', ST_GeomFromText('POLYGON((32.570 46.640, 32.590 46.640, 32.590 46.630, 32.570 46.630, 32.570 46.640))', 4326));

-- 4. Додаємо тестові самокати на вулиці
INSERT INTO vehicles (model_id, tariff_id, location, battery_level, status) VALUES
-- Самокати (Центр та Фабрика)
(1, 1, ST_SetSRID(ST_MakePoint(32.614600, 46.632200), 4326), 98, 'available'),
(2, 2, ST_SetSRID(ST_MakePoint(32.616200, 46.634500), 4326), 45, 'available'),
(2, 2, ST_SetSRID(ST_MakePoint(32.611100, 46.631500), 4326), 12, 'maintenance'),
(3, 3, ST_SetSRID(ST_MakePoint(32.643000, 46.673000), 4326), 100, 'available'),
(1, 3, ST_SetSRID(ST_MakePoint(32.644000, 46.674000), 4326), 88, 'available'),
-- Велосипеди (Шуменський та Таврійський)
(4, 4, ST_SetSRID(ST_MakePoint(32.580000, 46.635000), 4326), 100, 'available'),
(5, 5, ST_SetSRID(ST_MakePoint(32.620000, 46.665000), 4326), 90, 'available'),
(4, 4, ST_SetSRID(ST_MakePoint(32.615000, 46.635000), 4326), 100, 'available'),
-- Мопеди (Вокзал та Фабрика)
(6, 6, ST_SetSRID(ST_MakePoint(32.605000, 46.645000), 4326), 100, 'available'),
(6, 6, ST_SetSRID(ST_MakePoint(32.645000, 46.672000), 4326), 50, 'available'),
-- Моноколеса (Центр та Таврійський)
(7, 7, ST_SetSRID(ST_MakePoint(32.613000, 46.633000), 4326), 100, 'available'),
(8, 8, ST_SetSRID(ST_MakePoint(32.615500, 46.636000), 4326), 85, 'available'),
(7, 8, ST_SetSRID(ST_MakePoint(32.625000, 46.662000), 4326), 20, 'low_battery');

-- 5. Додаємо користувачів (Пароль: password123)
INSERT INTO users (name, email, phone, password_hash, role, balance, is_blocked) VALUES
('Адміністратор Системи', 'admin@example.com', '+380991112233', '$2a$10$5GWM0md73Y1cxH8L1K095epnSGewA7SSN2PeRNv7BFbn6SsULtdDG', 'admin', 0.00, false),
('Демо Користувач', 'demo@example.com', '+380991234567', '$2a$10$5GWM0md73Y1cxH8L1K095epnSGewA7SSN2PeRNv7BFbn6SsULtdDG', 'client', 500.00, false),
('Заблокований Юзер', 'blocked@example.com', '+380990001122', '$2a$10$5GWM0md73Y1cxH8L1K095epnSGewA7SSN2PeRNv7BFbn6SsULtdDG', 'client', 10.00, true);

-- 6. Додаємо історію поїздок для Демо Користувача
INSERT INTO rides (user_id, vehicle_id, status, start_time, end_time, start_location, end_location, total_price) VALUES 
(
    (SELECT id FROM users WHERE email = 'demo@example.com'), 1, 'completed', NOW() - INTERVAL '1 day' - INTERVAL '15 minutes', NOW() - INTERVAL '1 day', ST_SetSRID(ST_MakePoint(32.614600, 46.632200), 4326), ST_SetSRID(ST_MakePoint(32.611100, 46.631500), 4326), 54.00
);