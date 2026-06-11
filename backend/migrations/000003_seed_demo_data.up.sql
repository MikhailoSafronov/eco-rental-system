-- ДЕМО ДАНІ ДЛЯ ПРЕЗЕНТАЦІЇ ПРОЕКТУ
-- Цей файл створює готового користувача, щоб перевіряючий міг одразу тестувати систему

-- 1. Створюємо тестового користувача
-- Пароль зашифровано через bcrypt. Реальний пароль для входу: password123
INSERT INTO users (name, email, phone, password_hash, role, balance) VALUES
('Демо Користувач', 'demo@example.com', '+380991234567', '$2a$10$E97L/a6bQ0H1hs0GJndsi.coaXi2NV1EYeTpzP7FesjZzRtny513C', 'client', 500.00);