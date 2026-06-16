-- Очищаємо всі дані з таблиць, але залишаємо саму структуру бази
-- CASCADE потрібен, щоб ігнорувати зовнішні ключі (foreign keys) при очищенні

TRUNCATE TABLE rides CASCADE;
TRUNCATE TABLE vehicles CASCADE;
TRUNCATE TABLE vehicle_models CASCADE;
TRUNCATE TABLE tariffs CASCADE;
TRUNCATE TABLE parking_zones CASCADE;
TRUNCATE TABLE users CASCADE;