-- Видаляємо всі таблиці у зворотному порядку до їх створення
DROP TABLE IF EXISTS maintenance_logs CASCADE;
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS ride_telemetry CASCADE;
DROP TABLE IF EXISTS rides CASCADE;
DROP TABLE IF EXISTS vehicles CASCADE;
DROP TABLE IF EXISTS tariffs CASCADE;
DROP TABLE IF EXISTS vehicle_models CASCADE;
DROP TABLE IF EXISTS parking_zones CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Видаляємо типи
DROP TYPE IF EXISTS payment_status;
DROP TYPE IF EXISTS payment_type;
DROP TYPE IF EXISTS ride_status;
DROP TYPE IF EXISTS vehicle_status;
DROP TYPE IF EXISTS vehicle_type;
DROP TYPE IF EXISTS user_role;

-- Видаляємо розширення
DROP EXTENSION IF EXISTS postgis;
