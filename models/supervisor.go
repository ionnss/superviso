package models

import "time"

// Package models define as estruturas de dados principais do sistema.
//
// Contém:
//   - Modelos de usuário
//   - Modelos de supervisor
//   - Estruturas de agendamento

// Supervisor representa um profissional que oferece supervisão.
// Contém informações básicas do usuário e dados específicos de supervisor.
type Supervisor struct {
	UserID         int       `json:"user_id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	CRP            string    `json:"crp"`
	TheoryApproach string    `json:"theory_approach"`
	SessionPrice   float64   `json:"session_price"`
	AvailableDays  string    `json:"available_days"`
	StartTime      string    `json:"start_time"`
	EndTime        string    `json:"end_time"`
	CreatedAt      time.Time `json:"created_at"`
}