# Фитнес Клуб - Система управления

Курсовая работа: REST API сервер на Go с PostgreSQL и веб-клиентом.

## Структура проекта

```
.
├── backend/          # Go сервер
│   ├── main.go      # Точка входа
│   ├── models/      # Модели данных
│   ├── handlers/    # HTTP handlers
│   ├── database/    # Подключение к БД
│   └── middleware/  # Middleware (CORS)
├── frontend/        # HTML/CSS/JS клиент
│   ├── index.html
│   ├── style.css
│   └── script.js
└── migrations/      # SQL миграции
    └── schema.sql
```

## Установка и настройка

### 1. Установка PostgreSQL

Убедитесь, что PostgreSQL установлен и запущен:

```bash
# macOS (через Homebrew)
brew install postgresql
brew services start postgresql

# Проверьте статус
brew services list | grep postgresql
```

**Важно:** Если PostgreSQL уже запущен через `brew services`, НЕ нужно запускать его вручную в отдельном терминале!

### 2. Создание базы данных

```bash
# Создайте базу данных
createdb fitness_club

# Примените основную схему
psql -d fitness_club -f migrations/schema.sql

# Примените миграции для тренировок и авторизации
psql -d fitness_club -f migrations/add_trainings.sql
```

**Важно:** После применения миграций будет создан администратор по умолчанию:
- Email: `admin@fitness.club`
- Пароль: `admin`

### 3. Настройка backend

Файл `.env` уже создан в папке `backend/` с настройками по умолчанию.

**Важно:** Если у вас другой пользователь PostgreSQL (не `arkadiy`), отредактируйте `backend/.env`:
- `DB_USER` - имя вашего пользователя PostgreSQL (обычно это ваше имя пользователя в системе)
- `DB_PASSWORD` - пароль (обычно пустой для локального подключения через Homebrew)

### 4. Установка зависимостей Go

```bash
cd backend
go mod download
```

## Запуск

**Примечание:** Если PostgreSQL запущен через `brew services` (проверьте командой `brew services list`), то вам нужны только 2 терминала (backend и frontend), а не 3!

### Терминал 1: Backend сервер

```bash
cd backend
go run main.go
```

Сервер запустится на `http://localhost:8080`

### Терминал 2: Frontend клиент

```bash
cd frontend

# Используйте любой простой HTTP сервер, например:
python3 -m http.server 8000
# или
npx http-server -p 8000
```

Откройте в браузере: `http://localhost:8000`

## Система авторизации

### Вход в систему
- `POST /api/auth/login` - вход (возвращает токен)
- `POST /api/auth/logout` - выход
- `GET /api/auth/me` - получить текущего пользователя

**Учетные данные по умолчанию:**
- Администратор: `admin@fitness.club` / `admin`

Все защищенные endpoints требуют заголовок `Authorization` с токеном.

## API Endpoints

### Авторизация
- `POST /api/auth/login` - вход в систему
- `POST /api/auth/logout` - выход из системы
- `GET /api/auth/me` - получить текущего пользователя

### Пользователи (требует авторизации)
- `GET /api/users` - список всех пользователей
- `GET /api/users/{id}` - получить пользователя по ID
- `POST /api/users` - создать пользователя
- `DELETE /api/users/{id}` - удалить пользователя (только админ)

### Тренировки (требует авторизации)
- `GET /api/trainings` - список всех тренировок (с фильтрами: ?hall_type=, ?status=, ?trainer_id=)
- `GET /api/trainings/{id}` - получить тренировку по ID
- `POST /api/trainings` - создать тренировку (только тренер/админ)
- `PUT /api/trainings/{id}` - обновить тренировку (только создатель/админ)
- `DELETE /api/trainings/{id}` - удалить тренировку (только админ)
- `POST /api/trainings/{id}/register` - записаться на тренировку
- `POST /api/trainings/{id}/cancel` - отменить запись на тренировку

### Клиенты
- `GET /api/clients` - список всех клиентов
- `GET /api/clients/{id}` - получить клиента по ID
- `POST /api/clients` - создать клиента
- `DELETE /api/clients/{id}` - удалить клиента

### Абонементы
- `GET /api/subscriptions` - список всех абонементов
- `GET /api/subscriptions/{id}` - получить абонемент по ID
- `POST /api/subscriptions` - создать абонемент
- `DELETE /api/subscriptions/{id}` - удалить абонемент

### Сотрудники
- `GET /api/employees` - список всех сотрудников
- `GET /api/employees/{id}` - получить сотрудника по ID
- `POST /api/employees` - создать сотрудника
- `DELETE /api/employees/{id}` - удалить сотрудника

## Структура базы данных

- **users** - пользователи системы (user, trainer, admin)
- **sessions** - сессии авторизации
- **clients** - клиенты фитнес-клуба
- **subscriptions** - абонементы клиентов
- **employees** - сотрудники клуба
- **trainings** - тренировки (персональные и групповые)
- **training_participants** - участники тренировок

## Права доступа

### Администратор (admin)
- Полный доступ ко всем функциям
- Управление пользователями, клиентами, абонементами, сотрудниками
- Создание и удаление тренировок
- Просмотр всех данных

### Тренер (trainer)
- Создание тренировок (персональных и групповых)
- Редактирование своих тренировок
- Просмотр всех тренировок
- Просмотр участников своих тренировок

### Пользователь (user)
- Просмотр доступных тренировок
- Запись на тренировки
- Отмена записи на тренировки
- Просмотр своих тренировок

## Примечания

- Все логи выводятся в консоль backend сервера
- CORS настроен для работы с фронтендом
- Пароли хранятся в открытом виде (для курсовой работы достаточно)
- Токены авторизации действительны 24 часа
- Типы залов: pilates, yoga, gym, dance, cardio
- Типы тренировок: personal (персональная), group (групповая)
- Статусы тренировок: scheduled (запланировано), completed (завершено), cancelled (отменено)

