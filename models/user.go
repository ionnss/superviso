package models

import "time"

type User struct {
	ID             int       `json:"id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	CPF            string    `json:"cpf"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	CRP            string    `json:"crp"`
	TheoryApproach string    `json:"theory_approach"`
	CreatedAt      time.Time `json:"created_at"`
}
