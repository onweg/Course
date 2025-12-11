package middleware

import (
	"database/sql"
	"fitness-club/database"
	"log"
	"net/http"
)

// AuthMiddleware проверяет наличие валидного токена
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
			return
		}

		var userID int
		err := database.DB.QueryRow(`
			SELECT user_id 
			FROM sessions 
			WHERE token = $1 AND expires_at > NOW()
		`, token).Scan(&userID)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Недействительный или истекший токен", http.StatusUnauthorized)
				return
			}
			log.Printf("Ошибка проверки токена: %v", err)
			http.Error(w, "Ошибка проверки авторизации", http.StatusInternalServerError)
			return
		}

		// Сохраняем user_id в контексте запроса (можно использовать context в будущем)
		// Пока просто пропускаем запрос дальше
		next.ServeHTTP(w, r)
	})
}

// AdminOnly проверяет, что пользователь - администратор
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
			return
		}

		var role string
		err := database.DB.QueryRow(`
			SELECT u.role 
			FROM sessions s
			JOIN users u ON s.user_id = u.id
			WHERE s.token = $1 AND s.expires_at > NOW()
		`, token).Scan(&role)

		if err != nil || role != "admin" {
			http.Error(w, "Доступ запрещен. Требуются права администратора", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// TrainerOrAdmin проверяет, что пользователь - тренер или администратор
func TrainerOrAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
			return
		}

		var role string
		err := database.DB.QueryRow(`
			SELECT u.role 
			FROM sessions s
			JOIN users u ON s.user_id = u.id
			WHERE s.token = $1 AND s.expires_at > NOW()
		`, token).Scan(&role)

		if err != nil || (role != "trainer" && role != "admin") {
			http.Error(w, "Доступ запрещен. Требуются права тренера или администратора", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

