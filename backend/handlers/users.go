package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fitness-club/models"
	"fitness-club/database"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// GetUsers возвращает список всех пользователей
func GetUsers(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /api/users - получение списка пользователей")
	
	rows, err := database.DB.Query(`
		SELECT id, name, email, role, created_at 
		FROM users 
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt); err != nil {
			log.Printf("Ошибка сканирования: %v", err)
			continue
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
	log.Printf("Возвращено пользователей: %d", len(users))
}

// GetUser возвращает одного пользователя по ID
func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("GET /api/users/%d - получение пользователя", id)

	var u models.User
	err = database.DB.QueryRow(`
		SELECT id, name, email, role, created_at 
		FROM users 
		WHERE id = $1
	`, id).Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Пользователь не найден", http.StatusNotFound)
			return
		}
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// CreateUser создает нового пользователя
func CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("POST /api/users - создание пользователя")

	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Простая валидация
	if u.Name == "" || u.Email == "" {
		http.Error(w, "Имя и email обязательны", http.StatusBadRequest)
		return
	}

	// Проверяем уникальность email
	var existingEmail string
	err := database.DB.QueryRow(`
		SELECT email FROM users WHERE email = $1
	`, u.Email).Scan(&existingEmail)
	
	if err == nil {
		// Email уже существует
		http.Error(w, "Пользователь с таким email уже существует", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		// Другая ошибка при проверке
		log.Printf("Ошибка проверки email: %v", err)
		http.Error(w, "Ошибка проверки email", http.StatusInternalServerError)
		return
	}

	// Если пароль не указан, генерируем случайный
	if u.Password == "" {
		u.Password = generateRandomPassword()
		log.Printf("Сгенерирован случайный пароль для пользователя %s", u.Email)
	}

	if u.Role == "" {
		u.Role = "user"
	}

	var id int
	err = database.DB.QueryRow(`
		INSERT INTO users (name, email, password, role) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id
	`, u.Name, u.Email, u.Password, u.Role).Scan(&id)

	if err != nil {
		// Дополнительная проверка на случай, если уникальность нарушена между проверкой и вставкой
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "Пользователь с таким email уже существует", http.StatusConflict)
			return
		}
		log.Printf("Ошибка создания пользователя: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u.ID = id

	// Автоматически создаем связанные записи в зависимости от роли
	if u.Role == "user" {
		// Создаем клиента для обычного пользователя
		_, err = database.DB.Exec(`
			INSERT INTO clients (user_id) 
			VALUES ($1)
		`, id)
		if err != nil {
			log.Printf("Ошибка создания клиента: %v", err)
		} else {
			log.Printf("Автоматически создан клиент для пользователя %d", id)
		}
	} else if u.Role == "trainer" {
		// Создаем сотрудника для тренера
		_, err = database.DB.Exec(`
			INSERT INTO employees (user_id, position, hire_date) 
			VALUES ($1, $2, CURRENT_DATE)
		`, id, "Тренер")
		if err != nil {
			log.Printf("Ошибка создания сотрудника: %v", err)
		} else {
			log.Printf("Автоматически создан сотрудник для тренера %d", id)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
	log.Printf("Создан пользователь с ID: %d, роль: %s", id, u.Role)
}

// DeleteUser удаляет пользователя
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("DELETE /api/users/%d - удаление пользователя", id)

	result, err := database.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		log.Printf("Ошибка удаления: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Printf("Удален пользователь с ID: %d", id)
}

// generateRandomPassword генерирует случайный пароль
func generateRandomPassword() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Если не удалось сгенерировать случайный, используем простой
		return "password123"
	}
	return hex.EncodeToString(bytes)
}

