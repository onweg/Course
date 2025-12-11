package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"fitness-club/database"
	"fitness-club/models"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// GetTrainings возвращает список всех тренировок
func GetTrainings(w http.ResponseWriter, r *http.Request) {
	log.Println("GET /api/trainings - получение списка тренировок")

	// Получаем параметры фильтрации
	status := r.URL.Query().Get("status")
	hallType := r.URL.Query().Get("hall_type")
	trainerID := r.URL.Query().Get("trainer_id")

	query := `
		SELECT t.id, t.trainer_id, t.title, t.description, t.type, t.hall_type, 
		       t.start_time, t.duration_minutes, t.max_participants, t.current_participants, 
		       t.status, t.created_at,
		       u.id, u.name, u.email, u.role
		FROM trainings t
		LEFT JOIN users u ON t.trainer_id = u.id
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if status != "" {
		query += " AND t.status = $" + strconv.Itoa(argNum)
		args = append(args, status)
		argNum++
	}
	if hallType != "" {
		query += " AND t.hall_type = $" + strconv.Itoa(argNum)
		args = append(args, hallType)
		argNum++
	}
	if trainerID != "" {
		query += " AND t.trainer_id = $" + strconv.Itoa(argNum)
		args = append(args, trainerID)
		argNum++
	}

	query += " ORDER BY t.start_time ASC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var trainings []*models.Training
	trainingMap := make(map[int]*models.Training)
	
	for rows.Next() {
		var t models.Training
		var trainer models.User

		err := rows.Scan(&t.ID, &t.TrainerID, &t.Title, &t.Description, &t.Type, &t.HallType,
			&t.StartTime, &t.DurationMinutes, &t.MaxParticipants, &t.CurrentParticipants,
			&t.Status, &t.CreatedAt,
			&trainer.ID, &trainer.Name, &trainer.Email, &trainer.Role)
		if err != nil {
			log.Printf("Ошибка сканирования: %v", err)
			continue
		}

		t.Trainer = &trainer
		t.Participants = []models.TrainingParticipant{} // Инициализируем пустой массив
		trainingMap[t.ID] = &t
		trainings = append(trainings, &t)
	}
	
	// Загружаем участников для всех тренировок
	if len(trainingMap) > 0 {
		trainingIDs := make([]interface{}, 0, len(trainingMap))
		for id := range trainingMap {
			trainingIDs = append(trainingIDs, id)
		}
		
		// Строим запрос для загрузки участников
		placeholders := make([]string, len(trainingIDs))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}
		
		participantsQuery := fmt.Sprintf(`
			SELECT tp.id, tp.training_id, tp.user_id, tp.status, tp.registered_at,
			       u.id, u.name, u.email, u.role
			FROM training_participants tp
			JOIN users u ON tp.user_id = u.id
			WHERE tp.training_id IN (%s)
		`, strings.Join(placeholders, ","))
		
		participantsRows, err := database.DB.Query(participantsQuery, trainingIDs...)
		if err == nil {
			defer participantsRows.Close()
			for participantsRows.Next() {
				var p models.TrainingParticipant
				var u models.User
				err := participantsRows.Scan(&p.ID, &p.TrainingID, &p.UserID, &p.Status, &p.RegisteredAt,
					&u.ID, &u.Name, &u.Email, &u.Role)
				if err == nil {
					p.User = &u
					if training, exists := trainingMap[p.TrainingID]; exists {
						training.Participants = append(training.Participants, p)
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trainings)
	log.Printf("Возвращено тренировок: %d", len(trainings))
}

// GetTraining возвращает одну тренировку по ID с участниками
func GetTraining(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("GET /api/trainings/%d - получение тренировки", id)

	var t models.Training
	var trainer models.User

	err = database.DB.QueryRow(`
		SELECT t.id, t.trainer_id, t.title, t.description, t.type, t.hall_type, 
		       t.start_time, t.duration_minutes, t.max_participants, t.current_participants, 
		       t.status, t.created_at,
		       u.id, u.name, u.email, u.role
		FROM trainings t
		LEFT JOIN users u ON t.trainer_id = u.id
		WHERE t.id = $1
	`, id).Scan(&t.ID, &t.TrainerID, &t.Title, &t.Description, &t.Type, &t.HallType,
		&t.StartTime, &t.DurationMinutes, &t.MaxParticipants, &t.CurrentParticipants,
		&t.Status, &t.CreatedAt,
		&trainer.ID, &trainer.Name, &trainer.Email, &trainer.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Тренировка не найдена", http.StatusNotFound)
			return
		}
		log.Printf("Ошибка запроса: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Trainer = &trainer

	// Загружаем участников
	rows, err := database.DB.Query(`
		SELECT tp.id, tp.training_id, tp.user_id, tp.status, tp.registered_at,
		       u.id, u.name, u.email, u.role
		FROM training_participants tp
		JOIN users u ON tp.user_id = u.id
		WHERE tp.training_id = $1
	`, id)

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p models.TrainingParticipant
			var u models.User
			err := rows.Scan(&p.ID, &p.TrainingID, &p.UserID, &p.Status, &p.RegisteredAt,
				&u.ID, &u.Name, &u.Email, &u.Role)
			if err == nil {
				p.User = &u
				t.Participants = append(t.Participants, p)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// CreateTraining создает новую тренировку
func CreateTraining(w http.ResponseWriter, r *http.Request) {
	log.Println("POST /api/trainings - создание тренировки")

	// Получаем ID пользователя из токена
	token := r.Header.Get("Authorization")
	var currentUserID int
	var currentUserRole string
	err := database.DB.QueryRow(`
		SELECT u.id, u.role
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token = $1 AND s.expires_at > NOW()
	`, token).Scan(&currentUserID, &currentUserRole)

	if err != nil {
		http.Error(w, "Недействительный токен", http.StatusUnauthorized)
		return
	}

	var t models.Training
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		log.Printf("Ошибка декодирования: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Валидация
	if t.Title == "" || t.Type == "" || t.HallType == "" {
		http.Error(w, "Название, тип и тип зала обязательны", http.StatusBadRequest)
		return
	}

	// Проверяем trainer_id
	if t.TrainerID == 0 {
		// Если не указан, используем текущего пользователя (если он тренер или админ)
		if currentUserRole == "trainer" || currentUserRole == "admin" {
			t.TrainerID = currentUserID
		} else {
			http.Error(w, "Требуется указать тренера", http.StatusBadRequest)
			return
		}
	} else {
		// Проверяем, что указанный тренер существует и имеет роль trainer или admin
		var trainerRole string
		err = database.DB.QueryRow("SELECT role FROM users WHERE id = $1", t.TrainerID).Scan(&trainerRole)
		if err != nil {
			http.Error(w, "Тренер не найден", http.StatusBadRequest)
			return
		}
		if trainerRole != "trainer" && trainerRole != "admin" {
			http.Error(w, "Указанный пользователь не является тренером", http.StatusBadRequest)
			return
		}
	}

	if t.Type == "group" && t.MaxParticipants < 2 {
		http.Error(w, "Групповая тренировка должна иметь минимум 2 участника", http.StatusBadRequest)
		return
	}

	if t.Type == "personal" {
		t.MaxParticipants = 1
	}

	if t.DurationMinutes == 0 {
		t.DurationMinutes = 60
	}

	t.CurrentParticipants = 0
	if t.Status == "" {
		t.Status = "scheduled"
	}

	var id int
	err = database.DB.QueryRow(`
		INSERT INTO trainings (trainer_id, title, description, type, hall_type, start_time, 
		                       duration_minutes, max_participants, current_participants, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
		RETURNING id
	`, t.TrainerID, t.Title, t.Description, t.Type, t.HallType, t.StartTime,
		t.DurationMinutes, t.MaxParticipants, t.CurrentParticipants, t.Status).Scan(&id)

	if err != nil {
		log.Printf("Ошибка создания тренировки: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
	log.Printf("Создана тренировка с ID: %d", id)
}

// UpdateTraining обновляет тренировку
func UpdateTraining(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("PUT /api/trainings/%d - обновление тренировки", id)

	var t models.Training
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем права (только тренер-создатель или админ)
	token := r.Header.Get("Authorization")
	var userRole string
	var userID int
	var trainingTrainerID int
	err = database.DB.QueryRow(`
		SELECT u.role, u.id, t.trainer_id
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		JOIN trainings t ON t.id = $1
		WHERE s.token = $2 AND s.expires_at > NOW()
	`, id, token).Scan(&userRole, &userID, &trainingTrainerID)

	if err != nil {
		http.Error(w, "Недействительный токен", http.StatusUnauthorized)
		return
	}

	if userRole != "admin" && userID != trainingTrainerID {
		http.Error(w, "Доступ запрещен. Только создатель тренировки или администратор могут её редактировать", http.StatusForbidden)
		return
	}

	_, err = database.DB.Exec(`
		UPDATE trainings 
		SET title = $1, description = $2, type = $3, hall_type = $4, 
		    start_time = $5, duration_minutes = $6, max_participants = $7, status = $8
		WHERE id = $9
	`, t.Title, t.Description, t.Type, t.HallType, t.StartTime,
		t.DurationMinutes, t.MaxParticipants, t.Status, id)

	if err != nil {
		log.Printf("Ошибка обновления: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("Обновлена тренировка с ID: %d", id)
}

// // UpdateTrainingStatus позволяет тренеру-творцу или админу поменять статус тренировки
// func UpdateTrainingStatus(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	id, err := strconv.Atoi(vars["id"])
// 	if err != nil {
// 		http.Error(w, "Неверный ID", http.StatusBadRequest)
// 		return
// 	}

// 	type statusReq struct {
// 		Status string `json:"status"`
// 	}
// 	var req statusReq
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Неверный формат данных", http.StatusBadRequest)
// 		return
// 	}

// 	// Валидация статуса
// 	if req.Status != "scheduled" && req.Status != "completed" && req.Status != "cancelled" {
// 		http.Error(w, "Недопустимый статус", http.StatusBadRequest)
// 		return
// 	}

// 	// Получаем пользователя из токена и проверяем права
// 	token := r.Header.Get("Authorization")
// 	var userRole string
// 	var userID int
// 	var trainingTrainerID int
// 	err = database.DB.QueryRow(`
// 		SELECT u.role, u.id, t.trainer_id
// 		FROM sessions s
// 		JOIN users u ON s.user_id = u.id
// 		JOIN trainings t ON t.id = $1
// 		WHERE s.token = $2 AND s.expires_at > NOW()
// 	`, id, token).Scan(&userRole, &userID, &trainingTrainerID)

// 	if err != nil {
// 		http.Error(w, "Недействительный токен", http.StatusUnauthorized)
// 		return
// 	}

// 	if userRole != "admin" && userID != trainingTrainerID {
// 		http.Error(w, "Доступ запрещен. Только создатель тренировки или администратор могут менять статус", http.StatusForbidden)
// 		return
// 	}

// 	_, err = database.DB.Exec(`
// 		UPDATE trainings
// 		SET status = $1
// 		WHERE id = $2
// 	`, req.Status, id)

// 	if err != nil {
// 		log.Printf("Ошибка обновления статуса: %v", err)
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	log.Printf("Обновлен статус тренировки %d -> %s", id, req.Status)
// }

// DeleteTraining удаляет тренировку (только админ)
func DeleteTraining(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("DELETE /api/trainings/%d - удаление тренировки", id)

	result, err := database.DB.Exec("DELETE FROM trainings WHERE id = $1", id)
	if err != nil {
		log.Printf("Ошибка удаления: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Тренировка не найдена", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	log.Printf("Удалена тренировка с ID: %d", id)
}

// RegisterForTraining регистрирует пользователя на тренировку
func RegisterForTraining(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainingID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("POST /api/trainings/%d/register - регистрация на тренировку", trainingID)

	// Получаем ID пользователя из токена
	token := r.Header.Get("Authorization")
	var userID int
	var userRole string
	err = database.DB.QueryRow(`
		SELECT u.id, u.role
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.token = $1 AND s.expires_at > NOW()
	`, token).Scan(&userID, &userRole)

	if err != nil {
		http.Error(w, "Недействительный токен", http.StatusUnauthorized)
		return
	}

	// Если админ, может регистрировать другого пользователя через заголовок
	participantIDHeader := r.Header.Get("X-Participant-Id")
	if participantIDHeader != "" && userRole == "admin" {
		participantID, err := strconv.Atoi(participantIDHeader)
		if err == nil {
			userID = participantID
			log.Printf("Админ регистрирует пользователя %d на тренировку", userID)
		}
	}

	// Проверяем, есть ли место
	var training models.Training
	err = database.DB.QueryRow(`
		SELECT max_participants, current_participants, status, type
		FROM trainings 
		WHERE id = $1
	`, trainingID).Scan(&training.MaxParticipants, &training.CurrentParticipants, &training.Status, &training.Type)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Тренировка не найдена", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if training.Status != "scheduled" {
		http.Error(w, "Нельзя записаться на завершенную или отмененную тренировку", http.StatusBadRequest)
		return
	}

	if training.CurrentParticipants >= training.MaxParticipants {
		http.Error(w, "Нет свободных мест", http.StatusBadRequest)
		return
	}

	// Проверяем, не записан ли уже
	var exists int
	err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM training_participants 
		WHERE training_id = $1 AND user_id = $2
	`, trainingID, userID).Scan(&exists)

	if err == nil && exists > 0 {
		http.Error(w, "Вы уже записаны на эту тренировку", http.StatusBadRequest)
		return
	}

	// Проверяем, является ли пользователь клиентом с активным абонементом
	// Исключение: админы и тренеры могут записываться без абонемента
	if userRole != "admin" && userRole != "trainer" {
		var hasActiveSubscription bool
		err = database.DB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 
				FROM clients c
				JOIN subscriptions s ON c.id = s.client_id
				WHERE c.user_id = $1 
				AND s.status = 'active' 
				AND s.end_date >= CURRENT_DATE
			)
		`, userID).Scan(&hasActiveSubscription)

		if err != nil {
			log.Printf("Ошибка проверки абонемента: %v", err)
			http.Error(w, "Ошибка проверки абонемента", http.StatusInternalServerError)
			return
		}

		if !hasActiveSubscription {
			http.Error(w, "Для записи на тренировку необходим активный абонемент. Обратитесь к администратору для оформления абонемента.", http.StatusForbidden)
			return
		}
	}

	// Регистрируем
	_, err = database.DB.Exec(`
		INSERT INTO training_participants (training_id, user_id, status) 
		VALUES ($1, $2, 'registered')
	`, trainingID, userID)

	if err != nil {
		log.Printf("Ошибка регистрации: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Увеличиваем счетчик участников
	_, err = database.DB.Exec(`
		UPDATE trainings 
		SET current_participants = current_participants + 1 
		WHERE id = $1
	`, trainingID)

	if err != nil {
		log.Printf("Ошибка обновления счетчика: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	log.Printf("Пользователь %d зарегистрирован на тренировку %d", userID, trainingID)
}

// CancelRegistration отменяет регистрацию на тренировку
func CancelRegistration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trainingID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Ошибка парсинга ID: %v", err)
		http.Error(w, "Неверный ID", http.StatusBadRequest)
		return
	}

	log.Printf("POST /api/trainings/%d/cancel - отмена регистрации", trainingID)

	token := r.Header.Get("Authorization")
	if token == "" {
		log.Println("Токен не предоставлен")
		http.Error(w, "Токен не предоставлен", http.StatusUnauthorized)
		return
	}

	var userID int
	err = database.DB.QueryRow(`
		SELECT user_id 
		FROM sessions 
		WHERE token = $1 AND expires_at > NOW()
	`, token).Scan(&userID)

	if err != nil {
		log.Printf("Ошибка проверки токена: %v", err)
		http.Error(w, "Недействительный токен", http.StatusUnauthorized)
		return
	}

	log.Printf("Пользователь %d отменяет регистрацию на тренировку %d", userID, trainingID)

	result, err := database.DB.Exec(`
		DELETE FROM training_participants 
		WHERE training_id = $1 AND user_id = $2
	`, trainingID, userID)

	if err != nil {
		log.Printf("Ошибка отмены регистрации: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Регистрация не найдена", http.StatusNotFound)
		return
	}

	// Уменьшаем счетчик участников
	_, err = database.DB.Exec(`
		UPDATE trainings 
		SET current_participants = GREATEST(current_participants - 1, 0) 
		WHERE id = $1
	`, trainingID)

	w.WriteHeader(http.StatusOK)
	log.Printf("Регистрация пользователя %d на тренировку %d отменена", userID, trainingID)
}

// UpdateTrainingStatus позволяет тренеру или админу поменять статус тренировки
func UpdateTrainingStatus(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.Atoi(vars["id"])
    if err != nil {
        http.Error(w, "Неверный ID", http.StatusBadRequest)
        return
    }

    type statusReq struct {
        Status string `json:"status"`
    }
    var req statusReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Неверный формат данных", http.StatusBadRequest)
        return
    }

    // Валидация статуса
    if req.Status != "scheduled" && req.Status != "completed" && req.Status != "cancelled" {
        http.Error(w, "Недопустимый статус", http.StatusBadRequest)
        return
    }

    // Получаем пользователя из токена и проверяем права
    token := r.Header.Get("Authorization")
    var userRole string
    var userID int
    var trainingTrainerID int
    err = database.DB.QueryRow(`
        SELECT u.role, u.id, t.trainer_id
        FROM sessions s
        JOIN users u ON s.user_id = u.id
        JOIN trainings t ON t.id = $1
        WHERE s.token = $2 AND s.expires_at > NOW()
    `, id, token).Scan(&userRole, &userID, &trainingTrainerID)

    if err != nil {
        http.Error(w, "Недействительный токен", http.StatusUnauthorized)
        return
    }

    if userRole != "admin" && userID != trainingTrainerID {
        http.Error(w, "Доступ запрещен", http.StatusForbidden)
        return
    }

    _, err = database.DB.Exec(`UPDATE trainings SET status = $1 WHERE id = $2`, req.Status, id)
    if err != nil {
        log.Printf("Ошибка обновления статуса: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    log.Printf("Обновлен статус тренировки %d -> %s", id, req.Status)
}
