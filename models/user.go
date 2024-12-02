// superviso/models/user.go
package models

import "time"

// User represents a user in the system
type User struct {
	ID                   int       `json:"id"`
	FirstName            string    `json:"firstname"`
	LastName             string    `json:"lastname"`
	CPF                  string    `json:"cpf"`
	Email                string    `json:"email"`
	Password             string    `json:"-"` // Apenas para JSON
	PasswordHash         string    `json:"-"` // Armazenado no banco
	CRP                  string    `json:"crp"`
	TheoryApproach       string    `json:"approach"`
	Qualifications       string    `json:"qualifications"`
	UserRole             string    `json:"usertype"`
	SessionsAvailability string    `json:"availability"`
	CreatedAt            time.Time `json:"createdat"`
	FailedLoginAttempts  int       `json:"failed_login_attempts"`
	LastFailedLogin      time.Time `json:"last_failed_login"`
}

type SupervisorAvailability struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	AvailabilityDay  string    `json:"availability_day"`
	AvailabilityTime time.Time `json:"availability_time"`
	PricePerSession  float64   `json:"price_per_session"`
}
