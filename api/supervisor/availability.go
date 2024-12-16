package supervisor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"superviso/api/auth"
	"superviso/models"
)

// Criar novo arquivo para gerenciar disponibilidade
func UpdateWeeklyHours(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		// Processar cada dia da semana
		for day := 1; day <= 7; day++ {
			startTime := r.FormValue(fmt.Sprintf("start_time_%d", day))
			endTime := r.FormValue(fmt.Sprintf("end_time_%d", day))

			if startTime != "" && endTime != "" {
				_, err := db.Exec(`
					INSERT INTO supervisor_weekly_hours 
					(supervisor_id, weekday, start_time, end_time)
					VALUES ($1, $2, $3, $4)
					ON CONFLICT (supervisor_id, weekday) 
					DO UPDATE SET start_time = $3, end_time = $4`,
					userID, day, startTime, endTime)

				if err != nil {
					http.Error(w, "Erro ao atualizar horários", http.StatusInternalServerError)
					return
				}
			}
		}

		w.Write([]byte("Horários atualizados com sucesso"))
	}
}

// GetWeeklyHours retorna os horários semanais do supervisor
func GetWeeklyHours(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		supervisorID := r.URL.Query().Get("supervisor_id")
		if supervisorID == "" {
			http.Error(w, "ID do supervisor não fornecido", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(`
			SELECT weekday, start_time, end_time 
			FROM supervisor_weekly_hours 
			WHERE supervisor_id = $1 
			ORDER BY weekday`,
			supervisorID)
		if err != nil {
			http.Error(w, "Erro ao buscar horários", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var hours []models.SupervisorWeeklyHours
		for rows.Next() {
			var h models.SupervisorWeeklyHours
			err := rows.Scan(&h.Weekday, &h.StartTime, &h.EndTime)
			if err != nil {
				http.Error(w, "Erro ao ler horários", http.StatusInternalServerError)
				return
			}
			hours = append(hours, h)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hours)
	}
}

// CreateAvailabilityPeriod cria um novo período de disponibilidade
func CreateAvailabilityPeriod(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		var period models.SupervisorAvailabilityPeriod
		if err := json.NewDecoder(r.Body).Decode(&period); err != nil {
			http.Error(w, "Dados inválidos", http.StatusBadRequest)
			return
		}

		// Validar datas
		if period.StartDate.After(period.EndDate) {
			http.Error(w, "Data inicial deve ser anterior à data final", http.StatusBadRequest)
			return
		}

		_, err := db.Exec(`
			INSERT INTO supervisor_availability_periods 
			(supervisor_id, start_date, end_date)
			VALUES ($1, $2, $3)`,
			userID, period.StartDate, period.EndDate)

		if err != nil {
			http.Error(w, "Erro ao criar período", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

// GetAvailabilityPeriods retorna os períodos de disponibilidade do supervisor
func GetAvailabilityPeriods(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		supervisorID := r.URL.Query().Get("supervisor_id")
		if supervisorID == "" {
			http.Error(w, "ID do supervisor não fornecido", http.StatusBadRequest)
			return
		}

		rows, err := db.Query(`
			SELECT id, start_date, end_date, created_at 
			FROM supervisor_availability_periods 
			WHERE supervisor_id = $1 
			AND end_date >= CURRENT_DATE
			ORDER BY start_date`,
			supervisorID)
		if err != nil {
			http.Error(w, "Erro ao buscar períodos", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var periods []models.SupervisorAvailabilityPeriod
		for rows.Next() {
			var p models.SupervisorAvailabilityPeriod
			err := rows.Scan(&p.ID, &p.StartDate, &p.EndDate, &p.CreatedAt)
			if err != nil {
				http.Error(w, "Erro ao ler períodos", http.StatusInternalServerError)
				return
			}
			periods = append(periods, p)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(periods)
	}
}
