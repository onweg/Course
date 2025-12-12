package handlers

import (
	"database/sql"
	"encoding/json"
	"fitness-club/models"
	"fitness-club/database"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GetClients возвращает список всех клиентов
func GetClients(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /api/clients - получение списка клиентов")

	rows, err := database.DB.Query(`
		SELECT c.id, c.user_id, c.phone, c.address, c.birth_date, c.created_at,
		       u.id, u.name, u.email, u.role
		FROM clients c
		LEFT JOIN users u ON c.user_id = u.id
		ORDER BY c.created_at DESC
	`)
	if err != nil {
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var clients []models.Client
	// Инициализируем как пустой массив, а не nil
	clients = make([]models.Client, 0)
	
	for rows.Next() {
		var c models.Client
		var u models.User
		var birthDate sql.NullTime
		var phone sql.NullString
		var address sql.NullString

		err := rows.Scan(&c.ID, &c.UserID, &phone, &address, &birthDate, &c.CreatedAt,
			&u.ID, &u.Name, &u.Email, &u.Role)
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
		if birthDate.Valid {
			c.BirthDate = &birthDate.Time
		}
		c.User = &u
		clients = append(clients, c)
	}

	w.Header().Set("Content-Type", "application/json")
	// Всегда возвращаем массив, даже если пустой
	if clients == nil {
		clients = []models.Client{}
	}
	json.NewEncoder(w).Encode(clients)
	log.Printf("Возвращено клиентов: %d", len(clients))
}

// GetClient возвращает одного клиента по ID
func GetClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("GET /api/clients/%d - получение клиента", id)

	var c models.Client
	var u models.User
	var birthDate sql.NullTime
	var phone sql.NullString
	var address sql.NullString

	err = database.DB.QueryRow(`
		SELECT c.id, c.user_id, c.phone, c.address, c.birth_date, c.created_at,
		       u.id, u.name, u.email, u.role
		FROM clients c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.id = $1
	`, id).Scan(&c.ID, &c.UserID, &phone, &address, &birthDate, &c.CreatedAt,
		&u.ID, &u.Name, &u.Email, &u.Role)
	
	// Обрабатываем NULL значения
	if phone.Valid {
		c.Phone = phone.String
	}
	if address.Valid {
		c.Address = address.String
	}
	if birthDate.Valid {
		c.BirthDate = &birthDate.Time
	}

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Клиент не найден", http.StatusNotFound)
			return
		}
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if birthDate.Valid {
		c.BirthDate = &birthDate.Time
	}
	c.User = &u

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// CreateClient создает нового клиента
func CreateClient(w http.ResponseWriter, r *http.Request) {
	log.Println("POST /api/clients - создание клиента")

	var c models.Client
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if c.UserID == 0 {
		http.Error(w, "user_id обязателен", http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли пользователь
	var userExists bool
	err := database.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
	`, c.UserID).Scan(&userExists)
	
	if err != nil {
		log.Printf("Ошибка проверки пользователя: %v", err)
		http.Error(w, "Ошибка проверки пользователя", http.StatusInternalServerError)
		return
	}
	
	if !userExists {
		http.Error(w, "Пользователь с указанным ID не найден", http.StatusNotFound)
		return
	}

	// Проверяем, не существует ли уже клиент для этого пользователя
	var clientExists bool
	err = database.DB.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM clients WHERE user_id = $1)
	`, c.UserID).Scan(&clientExists)
	
	if err != nil {
		log.Printf("Ошибка проверки клиента: %v", err)
		http.Error(w, "Ошибка проверки клиента", http.StatusInternalServerError)
		return
	}
	
	if clientExists {
		http.Error(w, "Клиент для этого пользователя уже существует", http.StatusConflict)
		return
	}

	var id int
	var birthDate interface{}
	if c.BirthDate != nil {
		birthDate = *c.BirthDate
	}

	err = database.DB.QueryRow(`
		INSERT INTO clients (user_id, phone, address, birth_date) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id
	`, c.UserID, c.Phone, c.Address, birthDate).Scan(&id)

	if err != nil {
		log.Printf("Ошибка создания клиента: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
	log.Printf("Создан клиент с ID: %d", id)
}

// DeleteClient удаляет клиента
func DeleteClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("DELETE /api/clients/%d - удаление клиента", id)

	result, err := database.DB.Exec("DELETE FROM clients WHERE id = $1", id)
	if err != nil {
		log.Printf("Ошибка удаления: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Клиент не найден", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Printf("Удален клиент с ID: %d", id)
}

// UpdateClient обновляет данные клиента
func UpdateClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("PUT /api/clients/%d - обновление клиента", id)

	var c models.Client
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем существование клиента
	var exists bool
	err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM clients WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		log.Printf("Ошибка проверки клиента: %v", err)
		http.Error(w, "Ошибка проверки клиента", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Клиент не найден", http.StatusNotFound)
		return
	}

	// Обновляем поля
	var birthDate interface{}
	if c.BirthDate != nil {
		birthDate = *c.BirthDate
	}

	_, err = database.DB.Exec(`
		UPDATE clients 
		SET phone = $1, address = $2, birth_date = $3
		WHERE id = $4
	`, c.Phone, c.Address, birthDate, id)

	if err != nil {
		log.Printf("Ошибка обновления: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем обновленного клиента
	var updatedClient models.Client
	var u models.User
	var birthDateNull sql.NullTime
	var phone sql.NullString
	var address sql.NullString

	err = database.DB.QueryRow(`
		SELECT c.id, c.user_id, c.phone, c.address, c.birth_date, c.created_at,
		       u.id, u.name, u.email, u.role
		FROM clients c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.id = $1
	`, id).Scan(&updatedClient.ID, &updatedClient.UserID, &phone, &address, &birthDateNull, &updatedClient.CreatedAt,
		&u.ID, &u.Name, &u.Email, &u.Role)

	if err != nil {
		log.Printf("Ошибка получения обновленного клиента: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if phone.Valid {
		updatedClient.Phone = phone.String
	}
	if address.Valid {
		updatedClient.Address = address.String
	}
	if birthDateNull.Valid {
		updatedClient.BirthDate = &birthDateNull.Time
	}
	updatedClient.User = &u

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedClient)
	log.Printf("Обновлен клиент с ID: %d", id)
}

