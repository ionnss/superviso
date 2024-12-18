package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	tmpl "superviso/api/template"
	"time"
)

func GetScheduleHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		supervisorID := r.URL.Query().Get("supervisor_id")
		if supervisorID == "" {
			http.Error(w, "Supervisor ID não fornecido", http.StatusBadRequest)
			return
		}

		// Converter para inteiro
		supID, err := strconv.Atoi(supervisorID)
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		// Buscar informações do supervisor
		type SupervisorInfo struct {
			FirstName      string    `json:"first_name"`
			LastName       string    `json:"last_name"`
			CRP            string    `json:"crp"`
			TheoryApproach string    `json:"theory_approach"`
			SessionPrice   float64   `json:"session_price"`
			StartDate      time.Time `json:"start_date"`
			EndDate        time.Time `json:"end_date"`
			AvailableSlots []struct {
				SlotID    int       `json:"slot_id"`
				SlotDate  time.Time `json:"slot_date"`
				StartTime string    `json:"start_time"`
				EndTime   string    `json:"end_time"`
			} `json:"available_slots"`
		}

		var supervisor SupervisorInfo

		// Primeiro buscar informações básicas do supervisor
		err = db.QueryRow(`
			SELECT 
				u.first_name,
				u.last_name,
				u.crp, 
				u.theory_approach,
				sp.session_price,
				sap.start_date,
				sap.end_date
			FROM users u 
			JOIN supervisor_profiles sp ON u.id = sp.user_id 
			LEFT JOIN supervisor_availability_periods sap ON u.id = sap.supervisor_id
			WHERE u.id = $1`,
			supID).Scan(
			&supervisor.FirstName,
			&supervisor.LastName,
			&supervisor.CRP,
			&supervisor.TheoryApproach,
			&supervisor.SessionPrice,
			&supervisor.StartDate,
			&supervisor.EndDate,
		)

		if err != nil {
			log.Printf("Erro ao buscar supervisor: %v", err)
			http.Error(w, "Supervisor não encontrado", http.StatusNotFound)
			return
		}

		// Buscar slots disponíveis
		rows, err := db.Query(`
			SELECT 
				id,
				slot_date,
				start_time::text,
				end_time::text
			FROM available_slots 
			WHERE supervisor_id = $1 
			AND status = 'available'
			AND slot_date >= CURRENT_DATE
			ORDER BY slot_date, start_time`,
			supID)

		if err != nil {
			log.Printf("Erro ao buscar slots: %v", err)
			http.Error(w, "Erro ao buscar horários disponíveis", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var slot struct {
				SlotID    int       `json:"slot_id"`
				SlotDate  time.Time `json:"slot_date"`
				StartTime string    `json:"start_time"`
				EndTime   string    `json:"end_time"`
			}
			err := rows.Scan(&slot.SlotID, &slot.SlotDate, &slot.StartTime, &slot.EndTime)
			if err != nil {
				log.Printf("Erro ao ler slot: %v", err)
				continue
			}
			supervisor.AvailableSlots = append(supervisor.AvailableSlots, slot)
		}

		// Renderizar template
		tmpl := template.Must(template.New("schedule.html").
			Funcs(tmpl.TemplateFuncs).
			ParseFiles("view/schedule.html"))

		err = tmpl.Execute(w, map[string]interface{}{
			"Supervisor": supervisor,
		})

		if err != nil {
			log.Printf("Erro ao renderizar template: %v", err)
			http.Error(w, "Erro ao renderizar página", http.StatusInternalServerError)
			return
		}
	}
}
