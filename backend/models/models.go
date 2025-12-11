package models

import "time"

// User представляет пользователя системы
type User struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"password,omitempty" db:"password"`
	Role      string    `json:"role" db:"role"` // user, trainer, admin
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Client представляет клиента фитнес-клуба
type Client struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Phone     string    `json:"phone" db:"phone"`
	Address   string    `json:"address" db:"address"`
	BirthDate *time.Time `json:"birth_date,omitempty" db:"birth_date"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	User      *User     `json:"user,omitempty"`
}

// Subscription представляет абонемент
type Subscription struct {
	ID        int       `json:"id" db:"id"`
	ClientID  int       `json:"client_id" db:"client_id"`
	Type      string    `json:"type" db:"type"`
	StartDate time.Time `json:"start_date" db:"start_date"`
	EndDate   time.Time `json:"end_date" db:"end_date"`
	Price     float64   `json:"price" db:"price"`
	Status    string    `json:"status" db:"status"` // active, expired, cancelled
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Client    *Client   `json:"client,omitempty"`
}

// Employee представляет сотрудника
type Employee struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Position  string    `json:"position" db:"position"`
	Salary    *float64  `json:"salary,omitempty" db:"salary"`
	HireDate  time.Time `json:"hire_date" db:"hire_date"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	User      *User     `json:"user,omitempty"`
}

// Session представляет сессию пользователя
type Session struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	User      *User     `json:"user,omitempty"`
}

// Training представляет тренировку
type Training struct {
	ID                 int       `json:"id" db:"id"`
	TrainerID          int       `json:"trainer_id" db:"trainer_id"`
	Title              string    `json:"title" db:"title"`
	Description        string    `json:"description" db:"description"`
	Type               string    `json:"type" db:"type"` // personal, group
	HallType           string    `json:"hall_type" db:"hall_type"` // pilates, yoga, gym, dance, cardio
	StartTime          time.Time `json:"start_time" db:"start_time"`
	DurationMinutes    int       `json:"duration_minutes" db:"duration_minutes"`
	MaxParticipants    int       `json:"max_participants" db:"max_participants"`
	CurrentParticipants int      `json:"current_participants" db:"current_participants"`
	Status             string    `json:"status" db:"status"` // scheduled, completed, cancelled
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	Trainer            *User     `json:"trainer,omitempty"`
	Participants       []TrainingParticipant `json:"participants,omitempty"`
}

// TrainingParticipant представляет участника тренировки
type TrainingParticipant struct {
	ID           int       `json:"id" db:"id"`
	TrainingID   int       `json:"training_id" db:"training_id"`
	UserID       int       `json:"user_id" db:"user_id"`
	Status       string    `json:"status" db:"status"` // registered, attended, cancelled
	RegisteredAt time.Time `json:"registered_at" db:"registered_at"`
	User         *User     `json:"user,omitempty"`
}

// LoginRequest представляет запрос на вход
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse представляет ответ на вход
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

