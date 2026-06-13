-- Видаляємо історію поїздок для Демо Користувача
DELETE FROM rides WHERE user_id = (SELECT id FROM users WHERE email = 'demo@example.com');