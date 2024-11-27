// superviso/models/user.go
package models

import "time"

// User represents a user in the system
type User struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"firstname"`
	LastName     string    `json:"lastname"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Omissão em respostas json
	CRP          string    `json:"crp"`
	CPF          string    `json:"cpf"`
	Approach     string    `json:"approach"` // Abordagem clínica
	UserType     string    `json:"usertype"` // Supervisando ou Supervisor
	CreatedAt    time.Time `json:"createdat"`
	Active       bool      `json:"active"`
}
