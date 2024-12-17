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
	CreatedAt      time.Time `json:"created_at"`
}

type AvailabilityPeriod struct {
	ID           int       `json:"id"`
	SupervisorID int       `json:"supervisor_id"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	CreatedAt    time.Time `json:"created_at"`
}

// WeeklyHour representa um horário disponível em um dia da semana
type WeeklyHour struct {
	Weekday   int    `json:"weekday"` // 0-6 (Domingo-Sábado)
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}
