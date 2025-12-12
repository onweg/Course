package handlers

import (
	"database/sql"
	"encoding/json"
	"fitness-club/database"
	"log"
	"net/http"
)

// StatsResponse представляет статистику системы
type StatsResponse struct {
	TotalUsers       int     `json:"total_users"`
	TotalClients     int     `json:"total_clients"`
	TotalTrainers    int     `json:"total_trainers"`
	TotalTrainings   int     `json:"total_trainings"`
	ActiveSubscriptions int  `json:"active_subscriptions"`
	AverageTrainingDuration float64 `json:"average_training_duration"`
	UpcomingTrainings int   `json:"upcoming_trainings"`
	CompletedTrainings int  `json:"completed_trainings"`
}

// GetStats возвращает статистику системы
func GetStats(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /api/stats - получение статистики")

	stats := StatsResponse{}

	// Общее количество пользователей
	err := database.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		log.Printf("Ошибка подсчета пользователей: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Количество клиентов
	err = database.DB.QueryRow("SELECT COUNT(*) FROM clients").Scan(&stats.TotalClients)
	if err != nil {
		log.Printf("Ошибка подсчета клиентов: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Количество сотрудников (из таблицы employees)
	err = database.DB.QueryRow("SELECT COUNT(*) FROM employees").Scan(&stats.TotalTrainers)
	if err != nil {
		log.Printf("Ошибка подсчета сотрудников: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Общее количество тренировок
	err = database.DB.QueryRow("SELECT COUNT(*) FROM trainings").Scan(&stats.TotalTrainings)
	if err != nil {
		log.Printf("Ошибка подсчета тренировок: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Активные абонементы
	err = database.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM subscriptions 
		WHERE status = 'active' AND end_date >= CURRENT_DATE
	`).Scan(&stats.ActiveSubscriptions)
	if err != nil {
		log.Printf("Ошибка подсчета активных абонементов: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Средняя длительность тренировок (в минутах)
	var avgDuration sql.NullFloat64
	err = database.DB.QueryRow("SELECT AVG(duration_minutes) FROM trainings").Scan(&avgDuration)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Ошибка расчета средней длительности: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if avgDuration.Valid {
		stats.AverageTrainingDuration = avgDuration.Float64
	}

	// Предстоящие тренировки
	err = database.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM trainings 
		WHERE status = 'scheduled' AND start_time > NOW()
	`).Scan(&stats.UpcomingTrainings)
	if err != nil {
		log.Printf("Ошибка подсчета предстоящих тренировок: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Завершенные тренировки
	err = database.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM trainings 
		WHERE status = 'completed'
	`).Scan(&stats.CompletedTrainings)
	if err != nil {
		log.Printf("Ошибка подсчета завершенных тренировок: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
	log.Printf("Статистика возвращена успешно")
}

