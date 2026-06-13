-- Додаємо тестового адміністратора для перевірки адмін-панелі
-- Пароль зашифровано через bcrypt. Реальний пароль для входу: password123
INSERT INTO users (name, email, phone, password_hash, role, balance) VALUES
('Адміністратор Системи', 'admin@example.com', '+380991112233', '$2a$10$5GWM0md73Y1cxH8L1K095epnSGewA7SSN2PeRNv7BFbn6SsULtdDG', 'admin', 0.00);