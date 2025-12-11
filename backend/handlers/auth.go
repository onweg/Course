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

	log.Printf("Попытка входа: email=%s", req.Email)

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email и пароль обязательны", http.StatusBadRequest)
		return
	}

	// Проверяем пользователя
	var user models.User
	err := database.DB.QueryRow(`
		SELECT id, name, email, password, role, created_at 
		FROM users 
		WHERE email = $1
	`, req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
			return
		}
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Проверяем пароль (простая проверка, для курсовой достаточно)
	if user.Password != req.Password {
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

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

