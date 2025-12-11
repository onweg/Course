-- Добавление таблиц для тренировок и сессий

-- Таблица сессий для авторизации
CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица тренировок
CREATE TABLE IF NOT EXISTS trainings (
    id SERIAL PRIMARY KEY,
    trainer_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('personal', 'group')),
    hall_type VARCHAR(100) NOT NULL CHECK (hall_type IN ('pilates', 'yoga', 'gym', 'dance', 'cardio')),
    start_time TIMESTAMP NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 60,
    max_participants INTEGER DEFAULT 1,
    current_participants INTEGER DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'completed', 'cancelled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица участников тренировок
CREATE TABLE IF NOT EXISTS training_participants (
    id SERIAL PRIMARY KEY,
    training_id INTEGER REFERENCES trainings(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'registered' CHECK (status IN ('registered', 'attended', 'cancelled')),
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(training_id, user_id)
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
CREATE INDEX IF NOT EXISTS idx_trainings_trainer_id ON trainings(trainer_id);
CREATE INDEX IF NOT EXISTS idx_trainings_start_time ON trainings(start_time);
CREATE INDEX IF NOT EXISTS idx_training_participants_training_id ON training_participants(training_id);
CREATE INDEX IF NOT EXISTS idx_training_participants_user_id ON training_participants(user_id);

-- Создание администратора по умолчанию (если его еще нет)
INSERT INTO users (name, email, password, role) 
SELECT 'Администратор', 'admin@fitness.club', 'admin', 'admin'
WHERE NOT EXISTS (SELECT 1 FROM users WHERE email = 'admin@fitness.club');

