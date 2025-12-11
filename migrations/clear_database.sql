-- Скрипт для полной очистки базы данных
-- ВНИМАНИЕ: Этот скрипт удалит ВСЕ данные из всех таблиц!
-- Выполнить: psql -d fitness_club -f migrations/clear_database.sql

-- Отключаем проверку внешних ключей временно для более быстрого удаления
SET session_replication_role = 'replica';

-- Удаляем все данные из таблиц (в правильном порядке из-за внешних ключей)
TRUNCATE TABLE 
    training_participants,
    trainings,
    subscriptions,
    clients,
    employees,
    sessions,
    users
CASCADE;

-- Включаем обратно проверку внешних ключей
SET session_replication_role = 'origin';

-- Сбрасываем последовательности (автоинкременты) к началу
ALTER SEQUENCE IF EXISTS users_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS clients_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS subscriptions_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS employees_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS trainings_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS training_participants_id_seq RESTART WITH 1;
ALTER SEQUENCE IF EXISTS sessions_id_seq RESTART WITH 1;

-- Показываем результат
SELECT 'База данных очищена!' as status;

-- Проверяем, что таблицы пусты
SELECT 
    'users' as table_name, COUNT(*) as count FROM users
UNION ALL
SELECT 
    'clients', COUNT(*) FROM clients
UNION ALL
SELECT 
    'subscriptions', COUNT(*) FROM subscriptions
UNION ALL
SELECT 
    'employees', COUNT(*) FROM employees
UNION ALL
SELECT 
    'trainings', COUNT(*) FROM trainings
UNION ALL
SELECT 
    'training_participants', COUNT(*) FROM training_participants
UNION ALL
SELECT 
    'sessions', COUNT(*) FROM sessions;

