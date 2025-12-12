package main

import (
	"fitness-club/database"
	"fitness-club/handlers"
	"fitness-club/middleware"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Инициализация базы данных
	if err := database.InitDB(); err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer database.CloseDB()

	// Создание роутера
	r := mux.NewRouter()

	// Применение CORS middleware
	r.Use(middleware.CORS)

	// Универсальный обработчик OPTIONS для всех путей (должен быть до других маршрутов)
	r.Methods("OPTIONS").PathPrefix("/api").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Публичные маршруты (без авторизации)
	r.HandleFunc("/api/auth/login", handlers.Login).Methods("POST")

	// Защищенные маршруты (требуют авторизации)
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthMiddleware)

	// Авторизация
	api.HandleFunc("/auth/logout", handlers.Logout).Methods("POST")
	api.HandleFunc("/auth/me", handlers.GetCurrentUser).Methods("GET")

	// API маршруты для пользователей
	api.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	api.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	api.HandleFunc("/users/{id}", handlers.GetUser).Methods("GET")
	api.Handle("/users/{id}", middleware.AdminOnly(http.HandlerFunc(handlers.DeleteUser))).Methods("DELETE")

	// API маршруты для тренировок
	api.HandleFunc("/trainings/{id}/register", handlers.RegisterForTraining).Methods("POST")
	api.HandleFunc("/trainings/{id}/cancel", handlers.CancelRegistration).Methods("POST")
	api.Handle("/trainings/{id:[0-9]+}/status", middleware.TrainerOrAdmin(http.HandlerFunc(handlers.UpdateTrainingStatus))).Methods("PUT")
	api.HandleFunc("/trainings", handlers.GetTrainings).Methods("GET")
	api.Handle("/trainings", middleware.TrainerOrAdmin(http.HandlerFunc(handlers.CreateTraining))).Methods("POST")
	api.HandleFunc("/trainings/{id}", handlers.GetTraining).Methods("GET")
	api.Handle("/trainings/{id}", middleware.TrainerOrAdmin(http.HandlerFunc(handlers.UpdateTraining))).Methods("PUT")
	api.Handle("/trainings/{id}", middleware.AdminOnly(http.HandlerFunc(handlers.DeleteTraining))).Methods("DELETE")

	// API маршруты для клиентов
	api.HandleFunc("/clients", handlers.GetClients).Methods("GET")
	api.HandleFunc("/clients", handlers.CreateClient).Methods("POST")
	api.HandleFunc("/clients/{id}", handlers.GetClient).Methods("GET")
	api.Handle("/clients/{id}", middleware.AdminOnly(http.HandlerFunc(handlers.DeleteClient))).Methods("DELETE")

	// API маршруты для абонементов
	api.HandleFunc("/subscriptions", handlers.GetSubscriptions).Methods("GET")
	api.HandleFunc("/subscriptions", handlers.CreateSubscription).Methods("POST")
	api.HandleFunc("/subscriptions/{id}", handlers.GetSubscription).Methods("GET")
	api.Handle("/subscriptions/{id}", middleware.AdminOnly(http.HandlerFunc(handlers.DeleteSubscription))).Methods("DELETE")

	// API маршруты для сотрудников
	api.HandleFunc("/employees", handlers.GetEmployees).Methods("GET")
	api.HandleFunc("/employees", handlers.CreateEmployee).Methods("POST")
	api.HandleFunc("/employees/{id}", handlers.GetEmployee).Methods("GET")
	api.Handle("/employees/{id}", middleware.AdminOnly(http.HandlerFunc(handlers.DeleteEmployee))).Methods("DELETE")

	// Обработчик для несуществующих маршрутов (для отладки)
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Добавляем CORS заголовки даже для 404, чтобы браузер не ругался на CORS
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Participant-Id, Accept, Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type, Authorization")

		log.Printf("404: Маршрут не найден: %s %s", r.Method, r.URL.Path)
		http.Error(w, "Маршрут не найден: "+r.Method+" "+r.URL.Path, http.StatusNotFound)
	})

	// Получение порта из переменных окружения
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Сервер запущен на порту %s", port)
	log.Printf("API доступен по адресу: http://localhost:%s/api", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

