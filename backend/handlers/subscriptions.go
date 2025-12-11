package handlers

import (
	"database/sql"
	"encoding/json"
	"fitness-club/models"
	"fitness-club/database"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// GetSubscriptions возвращает список всех абонементов
func GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /api/subscriptions - получение списка абонементов")

	rows, err := database.DB.Query(`
		SELECT s.id, s.client_id, s.type, s.start_date, s.end_date, s.price, s.status, s.created_at,
		       c.id, c.user_id, c.phone, c.address
		FROM subscriptions s
		LEFT JOIN clients c ON s.client_id = c.id
		ORDER BY s.created_at DESC
	`)
	if err != nil {
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var subscriptions []models.Subscription
	// Инициализируем как пустой массив
	subscriptions = make([]models.Subscription, 0)
	
	for rows.Next() {
		var s models.Subscription
		var c models.Client
		var phone sql.NullString
		var address sql.NullString

		err := rows.Scan(&s.ID, &s.ClientID, &s.Type, &s.StartDate, &s.EndDate, &s.Price, &s.Status, &s.CreatedAt,
			&c.ID, &c.UserID, &phone, &address)
		if err != nil {
			log.Printf("Ошибка сканирования: %v", err)
			continue
		}

		// Обрабатываем NULL значения
		if phone.Valid {
			c.Phone = phone.String
		}
		if address.Valid {
			c.Address = address.String
		}
		
		// Автоматически обновляем статус на основе текущей даты
		now := time.Now()
		originalStatus := s.Status
		if s.Status != "cancelled" {
			if s.EndDate.Before(now) {
				s.Status = "expired"
			} else if s.StartDate.Before(now) || s.StartDate.Equal(now) {
				s.Status = "active"
			}
			
			// Обновляем статус в БД, если он изменился
			if s.Status != originalStatus {
				_, _ = database.DB.Exec(`
					UPDATE subscriptions 
					SET status = $1 
					WHERE id = $2
				`, s.Status, s.ID)
			}
		}
		
		s.Client = &c
		subscriptions = append(subscriptions, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subscriptions)
	log.Printf("Возвращено абонементов: %d", len(subscriptions))
}

// GetSubscription возвращает один абонемент по ID
func GetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("GET /api/subscriptions/%d - получение абонемента", id)

	var s models.Subscription
	var c models.Client

	err = database.DB.QueryRow(`
		SELECT s.id, s.client_id, s.type, s.start_date, s.end_date, s.price, s.status, s.created_at,
		       c.id, c.user_id, c.phone, c.address
		FROM subscriptions s
		LEFT JOIN clients c ON s.client_id = c.id
		WHERE s.id = $1
	`, id).Scan(&s.ID, &s.ClientID, &s.Type, &s.StartDate, &s.EndDate, &s.Price, &s.Status, &s.CreatedAt,
		&c.ID, &c.UserID, &c.Phone, &c.Address)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Абонемент не найден", http.StatusNotFound)
			return
		}
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.Client = &c

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s)
}

// SubscriptionRequest представляет запрос на создание абонемента (с датами в виде строк)
type SubscriptionRequest struct {
	UserID    int    `json:"user_id"`    // ID пользователя (будем искать его client_id)
	Type      string `json:"type"`       // Тип абонемента: monthly, quarterly, yearly
	StartDate string `json:"start_date"` // Формат: YYYY-MM-DD
}

// CreateSubscription создает новый абонемент
func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	log.Println("POST /api/subscriptions - создание абонемента")

	var req SubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Ошибка декодирования JSON: %v", err)
		http.Error(w, "Неверный формат данных: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Получены данные абонемента: UserID=%d, Type=%s, StartDate=%s",
		req.UserID, req.Type, req.StartDate)

	// Валидация обязательных полей
	if req.UserID == 0 {
		log.Println("Ошибка: user_id не указан или равен 0")
		http.Error(w, "user_id обязателен и должен быть больше 0", http.StatusBadRequest)
		return
	}
	
	if req.Type == "" {
		log.Println("Ошибка: type не указан")
		http.Error(w, "type обязателен", http.StatusBadRequest)
		return
	}
	
	// Находим client_id по user_id
	var clientID int
	err := database.DB.QueryRow(`
		SELECT id FROM clients WHERE user_id = $1
	`, req.UserID).Scan(&clientID)
	
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Ошибка: клиент для пользователя %d не найден", req.UserID)
			http.Error(w, "Клиент для этого пользователя не найден. Убедитесь, что пользователь имеет роль 'user'", http.StatusNotFound)
			return
		}
		log.Printf("Ошибка поиска клиента: %v", err)
		http.Error(w, "Ошибка поиска клиента", http.StatusInternalServerError)
		return
	}
	
	// Парсим дату начала
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		log.Printf("Ошибка парсинга start_date '%s': %v", req.StartDate, err)
		http.Error(w, "Неверный формат start_date. Ожидается YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	
	// Рассчитываем дату окончания и цену на основе типа абонемента
	var endDate time.Time
	var price float64
	var typeName string
	
	switch req.Type {
	case "monthly", "Месячный":
		endDate = startDate.AddDate(0, 1, 0) // +1 месяц
		price = 2000.0
		typeName = "Месячный"
	case "quarterly", "Квартальный":
		endDate = startDate.AddDate(0, 3, 0) // +3 месяца
		price = 5000.0
		typeName = "Квартальный"
	case "yearly", "Годовой":
		endDate = startDate.AddDate(1, 0, 0) // +1 год
		price = 18000.0
		typeName = "Годовой"
	default:
		log.Printf("Ошибка: неизвестный тип абонемента '%s'", req.Type)
		http.Error(w, "Неизвестный тип абонемента. Доступны: monthly, quarterly, yearly", http.StatusBadRequest)
		return
	}
	
	// Автоматически определяем статус на основе текущей даты
	now := time.Now()
	var status string
	if endDate.Before(now) {
		status = "expired"
	} else if startDate.After(now) {
		status = "active" // Абонемент еще не начался, но активен
	} else {
		status = "active" // Абонемент активен
	}
	
	// Создаем модель Subscription
	s := models.Subscription{
		ClientID:  clientID,
		Type:      typeName,
		StartDate: startDate,
		EndDate:   endDate,
		Price:     price,
		Status:    status,
	}

	var id int
	err = database.DB.QueryRow(`
		INSERT INTO subscriptions (client_id, type, start_date, end_date, price, status) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id
	`, s.ClientID, s.Type, s.StartDate, s.EndDate, s.Price, s.Status).Scan(&id)

	if err != nil {
		log.Printf("Ошибка создания абонемента: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
	log.Printf("Создан абонемент с ID: %d", id)
}

// DeleteSubscription удаляет абонемент
func DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("DELETE /api/subscriptions/%d - удаление абонемента", id)

	result, err := database.DB.Exec("DELETE FROM subscriptions WHERE id = $1", id)
	if err != nil {
		log.Printf("Ошибка удаления: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Абонемент не найден", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Printf("Удален абонемент с ID: %d", id)
}

