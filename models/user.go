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
	PasswordHash         string    `json:"-"` // Omissão em respostas json
	CRP                  string    `json:"crp"`
	TheoryApproach       string    `json:"approach"`          // Abordagem teórica
	Qualifications       string    `json:"qualifications"`    // Qualificações do usuário
	UserRole             string    `json:"usertype"`          // Supervisando ou Supervisor
	PricePerSession      float64   `json:"price_per_session"` // Preço por sessão
	SessionsAvailability string    `json:"availability"`      // Disponibilidade
	CreatedAt            time.Time `json:"createdat"`
}
