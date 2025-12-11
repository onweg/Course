-- Скрипт для создания клиентов для существующих пользователей с ролью 'user'
-- Выполнить: psql -d fitness_club -f migrations/create_missing_clients.sql

-- Создаем клиентов для всех пользователей с ролью 'user', у которых еще нет клиента
INSERT INTO clients (user_id)
SELECT u.id
FROM users u
WHERE u.role = 'user'
  AND NOT EXISTS (
    SELECT 1 FROM clients c WHERE c.user_id = u.id
  );

-- Показываем результат
SELECT 
    u.id as user_id,
    u.name as user_name,
    u.email,
    c.id as client_id
FROM users u
LEFT JOIN clients c ON u.id = c.user_id
WHERE u.role = 'user'
ORDER BY u.id;

