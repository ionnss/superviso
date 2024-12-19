package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type AvailableSlot struct {
	ID           int       `json:"id"`
	SupervisorID int       `json:"supervisor_id"`
	Date         time.Time `json:"date"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

func GetAvailableSlots(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `
			SELECT 
				id,
				supervisor_id,
				slot_date,
				start_time,
				end_time
			FROM available_slots
			WHERE slot_date >= CURRENT_DATE 
			AND status = 'available'
			ORDER BY slot_date, start_time`

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Erro ao buscar slots", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var slots []AvailableSlot
		for rows.Next() {
			var slot AvailableSlot
			err := rows.Scan(
				&slot.ID,
				&slot.SupervisorID,
				&slot.Date,
				&slot.StartTime,
				&slot.EndTime,
			)
			if err != nil {
				http.Error(w, "Erro ao ler slots", http.StatusInternalServerError)
				return
			}
			slots = append(slots, slot)
		}
		if err = rows.Err(); err != nil {
			http.Error(w, "Erro ao ler slots", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(slots); err != nil {
			http.Error(w, "Erro ao codificar resposta", http.StatusInternalServerError)
			return
		}
	}
}
