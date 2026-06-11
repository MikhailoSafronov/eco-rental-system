-- Додаємо одну завершену поїздку в історію для Демо Користувача
INSERT INTO rides (user_id, vehicle_id, status, start_time, end_time, start_location, end_location, total_price) 
VALUES (
    (SELECT id FROM users WHERE email = 'demo@example.com'), 
    1, 'completed', NOW() - INTERVAL '1 day' - INTERVAL '15 minutes', NOW() - INTERVAL '1 day',
    ST_SetSRID(ST_MakePoint(32.614600, 46.632200), 4326), ST_SetSRID(ST_MakePoint(32.611100, 46.631500), 4326), 54.00
);