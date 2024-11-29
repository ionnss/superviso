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
	PricePerSession      float64   `json:"price_per_session"`
	SessionsAvailability string    `json:"availability"`
	CreatedAt            time.Time `json:"createdat"`
}
