package middleware

import (
	"log"
	"net/http"
)

// CORS middleware для разрешения запросов с фронтенда
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CORS: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Participant-Id")

		// Обрабатываем preflight OPTIONS запросы
		if r.Method == "OPTIONS" {
			log.Printf("CORS: обработан OPTIONS запрос для %s", r.URL.Path)
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

