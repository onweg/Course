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

// GetEmployees возвращает список всех сотрудников
func GetEmployees(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /api/employees - получение списка сотрудников")

	rows, err := database.DB.Query(`
		SELECT e.id, e.user_id, e.position, e.salary, e.hire_date, e.created_at,
		       u.id, u.name, u.email, u.role
		FROM employees e
		LEFT JOIN users u ON e.user_id = u.id
		ORDER BY e.created_at DESC
	`)
	if err != nil {
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var e models.Employee
		var u models.User
		var salary sql.NullFloat64

		err := rows.Scan(&e.ID, &e.UserID, &e.Position, &salary, &e.HireDate, &e.CreatedAt,
			&u.ID, &u.Name, &u.Email, &u.Role)
		if err != nil {
			log.Printf("Ошибка сканирования: %v", err)
			continue
		}

		if salary.Valid {
			e.Salary = &salary.Float64
		}
		e.User = &u
		employees = append(employees, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(employees)
	log.Printf("Возвращено сотрудников: %d", len(employees))
}

// GetEmployee возвращает одного сотрудника по ID
func GetEmployee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("GET /api/employees/%d - получение сотрудника", id)

	var e models.Employee
	var u models.User
	var salary sql.NullFloat64

	err = database.DB.QueryRow(`
		SELECT e.id, e.user_id, e.position, e.salary, e.hire_date, e.created_at,
		       u.id, u.name, u.email, u.role
		FROM employees e
		LEFT JOIN users u ON e.user_id = u.id
		WHERE e.id = $1
	`, id).Scan(&e.ID, &e.UserID, &e.Position, &salary, &e.HireDate, &e.CreatedAt,
		&u.ID, &u.Name, &u.Email, &u.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Сотрудник не найден", http.StatusNotFound)
			return
		}
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if salary.Valid {
		e.Salary = &salary.Float64
	}
	e.User = &u

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

// CreateEmployee создает нового сотрудника
func CreateEmployee(w http.ResponseWriter, r *http.Request) {
	log.Println("POST /api/employees - создание сотрудника")

	var e models.Employee
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if e.UserID == 0 || e.Position == "" {
		http.Error(w, "user_id и position обязательны", http.StatusBadRequest)
		return
	}

	var id int
	var salary interface{}
	if e.Salary != nil {
		salary = *e.Salary
	}

	err := database.DB.QueryRow(`
		INSERT INTO employees (user_id, position, salary, hire_date) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id
	`, e.UserID, e.Position, salary, e.HireDate).Scan(&id)

	if err != nil {
		log.Printf("Ошибка создания сотрудника: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	e.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
	log.Printf("Создан сотрудник с ID: %d", id)
}

// DeleteEmployee удаляет сотрудника
func DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("DELETE /api/employees/%d - удаление сотрудника", id)

	result, err := database.DB.Exec("DELETE FROM employees WHERE id = $1", id)
	if err != nil {
		log.Printf("Ошибка удаления: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Сотрудник не найден", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Printf("Удален сотрудник с ID: %d", id)
}

