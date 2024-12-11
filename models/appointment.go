package models

import (
	"time"
)

type AvailableSlot struct {
	ID           int       `json:"id"`
	SupervisorID int       `json:"supervisor_id"`
	SlotDate     time.Time `json:"slot_date"`
	StartTime    string    `json:"start_time"`
	EndTime      string    `json:"end_time"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Appointment struct {
	ID                 int       `json:"id"`
	SupervisorID       int       `json:"supervisor_id"`
	SuperviseeID       int       `json:"supervisee_id"`
	SlotID             int       `json:"slot_id"`
	Status             string    `json:"status"`
	CancellationReason string    `json:"cancellation_reason,omitempty"`
	Notes              string    `json:"notes,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
