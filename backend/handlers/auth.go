package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fitness-club/database"
	"fitness-club/models"
	"log"
	"net/http"
	"strings"
	"time"
)

// Login обрабатывает вход пользователя
func Login(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST /api/auth/login - вход пользователя (Method: %s, RemoteAddr: %s)", r.Method, r.RemoteAddr)

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Ошибка декодирования запроса: %v", err)
		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
		return
	}

	// Нормализуем email (убираем пробелы, приводим к нижнему регистру)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)
	
	log.Printf("Попытка входа: email='%s' (длина=%d), password='%s' (длина=%d)", 
		req.Email, len(req.Email), req.Password, len(req.Password))

	if req.Email == "" || req.Password == "" {
		log.Printf("Пустой email или пароль: email='%s', password='%s'", req.Email, req.Password)
		http.Error(w, "Email и пароль обязательны", http.StatusBadRequest)
		return
	}

	// Проверяем пользователя
	var user models.User
	var passwordFromDB string
	err := database.DB.QueryRow(`
		SELECT id, name, email, password, role, created_at 
		FROM users 
		WHERE LOWER(TRIM(email)) = LOWER(TRIM($1))
	`, req.Email).Scan(&user.ID, &user.Name, &user.Email, &passwordFromDB, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Пользователь с email '%s' не найден в базе данных", req.Email)
			// Проверяем, может быть похожий email существует
			var similarEmail string
			checkErr := database.DB.QueryRow(`
				SELECT email FROM users WHERE LOWER(email) LIKE LOWER($1) LIMIT 1
			`, "%"+strings.Split(req.Email, "@")[0]+"%").Scan(&similarEmail)
			if checkErr == nil {
				log.Printf("Найден похожий email в базе: %s (возможно опечатка)", similarEmail)
			}
			http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
			return
		}
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Нормализуем пароль из БД (убираем пробелы)
	passwordFromDB = strings.TrimSpace(passwordFromDB)
	
	// Проверяем, что пароль не пустой в базе данных
	if passwordFromDB == "" {
		log.Printf("Пользователь %s (ID: %d) имеет пустой пароль в базе данных", req.Email, user.ID)
		http.Error(w, "Пароль пользователя не установлен. Обратитесь к администратору", http.StatusUnauthorized)
		return
	}

	user.Password = passwordFromDB

	// Проверяем пароль (простая проверка, для курсовой достаточно)
	log.Printf("Проверка пароля для пользователя %s (ID: %d). Пароль из БД: [%s] (длина=%d), введенный пароль: [%s] (длина=%d)", 
		req.Email, user.ID, user.Password, len(user.Password), req.Password, len(req.Password))
	
	if user.Password != req.Password {
		log.Printf("Неверный пароль для пользователя %s (ID: %d). Ожидался: [%s] (байты: %v), получен: [%s] (байты: %v)", 
			req.Email, user.ID, user.Password, []byte(user.Password), req.Password, []byte(req.Password))
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}
	
	log.Printf("Пароль совпадает для пользователя %s (ID: %d)", req.Email, user.ID)

	// Генерируем токен
	token, err := generateToken()
	if err != nil {
		log.Printf("Ошибка генерации токена: %v", err)
		http.Error(w, "Ошибка создания сессии", http.StatusInternalServerError)
		return
	}

	// Сохраняем сессию (действует 24 часа)
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = database.DB.Exec(`
		INSERT INTO sessions (user_id, token, expires_at) 
		VALUES ($1, $2, $3)
	`, user.ID, token, expiresAt)

	if err != nil {
		log.Printf("Ошибка создания сессии: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Удаляем пароль из ответа
	user.Password = ""

	response := models.LoginResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("Пользователь %s (%s) вошел в систему", user.Name, user.Role)
}

// Logout обрабатывает выход пользователя
func Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Токен не предоставлен", http.StatusBadRequest)
		return
	}

	// Удаляем сессию
	_, err := database.DB.Exec("DELETE FROM sessions WHERE token = $1", token)
	if err != nil {
		log.Printf("Ошибка удаления сессии: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	log.Println("Пользователь вышел из системы")
}

// GetCurrentUser возвращает текущего пользователя по токену
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
		return
	}

	var user models.User
	var sessionID int
	err := database.DB.QueryRow(`
		SELECT s.id, u.id, u.name, u.email, u.role, u.created_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token = $1 AND s.expires_at > NOW()
	`, token).Scan(&sessionID, &user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Недействительный или истекший токен", http.StatusUnauthorized)
			return
		}
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// generateToken генерирует случайный токен
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Register обрабатывает регистрацию нового пользователя
func Register(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name     string `json:"name"`
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Некорректные данные", http.StatusBadRequest)
        return
    }
    // Базовая проверка
    req.Name = strings.TrimSpace(req.Name)
    req.Email = strings.TrimSpace(req.Email)
    req.Password = strings.TrimSpace(req.Password)
    if req.Name == "" || req.Email == "" || req.Password == "" {
        http.Error(w, "Все поля обязательны", http.StatusBadRequest)
        return
    }
    // Проверка уникальности email
    var exists int
    err := database.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE email = $1`, req.Email).Scan(&exists)
    if err != nil || exists > 0 {
        http.Error(w, "Пользователь с таким email уже существует", http.StatusConflict)
        return
    }
    // Создание пользователя
    var id int
    err = database.DB.QueryRow(`
        INSERT INTO users (name, email, password, role)
        VALUES ($1, $2, $3, 'user') RETURNING id
    `, req.Name, req.Email, req.Password).Scan(&id)
    if err != nil {
        http.Error(w, "Ошибка создания пользователя", http.StatusInternalServerError)
        return
    }

    // Загружаем созданного пользователя
    var user models.User
    err = database.DB.QueryRow(`
        SELECT id, name, email, role, created_at
        FROM users
        WHERE id = $1
    `, id).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt)
    if err != nil {
        http.Error(w, "Ошибка чтения созданного пользователя", http.StatusInternalServerError)
        return
    }

    // Генерируем токен и создаем сессию, как при логине
    token, err := generateToken()
    if err != nil {
        log.Printf("Ошибка генерации токена при регистрации: %v", err)
        http.Error(w, "Ошибка создания сессии", http.StatusInternalServerError)
        return
    }

    expiresAt := time.Now().Add(24 * time.Hour)
    _, err = database.DB.Exec(`
        INSERT INTO sessions (user_id, token, expires_at)
        VALUES ($1, $2, $3)
    `, user.ID, token, expiresAt)
    if err != nil {
        log.Printf("Ошибка создания сессии при регистрации: %v", err)
        http.Error(w, "Ошибка создания сессии", http.StatusInternalServerError)
        return
    }

    // Формируем такой же ответ, как при логине
    response := models.LoginResponse{
        Token: token,
        User:  user,
    }

    w.Header().Set("Content-Type", "application/json")
    // Можно оставить 201, чтобы явно обозначить создание
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}