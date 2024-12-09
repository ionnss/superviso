package models

import "time"

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
