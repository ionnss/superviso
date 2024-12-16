package supervisor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"superviso/api/auth"
	"superviso/models"

	"github.com/gorilla/mux"
)

// Criar novo arquivo para gerenciar disponibilidade
func UpdateWeeklyHours(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		// Iniciar transação
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Erro ao iniciar transação", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Limpar horários existentes
		_, err = tx.Exec(`DELETE FROM supervisor_weekly_hours WHERE supervisor_id = $1`, userID)
		if err != nil {
			http.Error(w, "Erro ao limpar horários", http.StatusInternalServerError)
			return
		}

		// Inserir novos horários
		for day := 1; day <= 7; day++ {
			startTime := r.FormValue(fmt.Sprintf("start_time_%d", day))
			endTime := r.FormValue(fmt.Sprintf("end_time_%d", day))

			if startTime != "" && endTime != "" {
				_, err = tx.Exec(`
					INSERT INTO supervisor_weekly_hours 
					(supervisor_id, weekday, start_time, end_time)
					VALUES ($1, $2, $3, $4)`,
					userID, day, startTime, endTime)

				if err != nil {
					http.Error(w, "Erro ao salvar horários", http.StatusInternalServerError)
					return
				}
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Erro ao finalizar alterações", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(`<div class="alert alert-success">Horários atualizados com sucesso!</div>`))
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
		startDate := r.FormValue("availability_start")
		endDate := r.FormValue("availability_end")

		_, err := db.Exec(`
			INSERT INTO supervisor_availability_periods 
			(supervisor_id, start_date, end_date)
			 VALUES ($1, $2, $3)`,
			userID, startDate, endDate)

		if err != nil {
			http.Error(w, "Erro ao criar período", http.StatusInternalServerError)
			return
		}

		// Retornar HTML do novo período
		tmpl := template.Must(template.ParseFiles("view/partials/availability_period.html"))
		tmpl.Execute(w, map[string]interface{}{
			"StartDate": startDate,
			"EndDate":   endDate,
		})
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

// Adicionar esta função
func DeleteAvailabilityPeriod(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)
		periodID := mux.Vars(r)["id"]

		result, err := db.Exec(`
			DELETE FROM supervisor_availability_periods 
			WHERE id = $1 AND supervisor_id = $2`,
			periodID, userID)

		if err != nil {
			http.Error(w, "Erro ao deletar período", http.StatusInternalServerError)
			return
		}

		rows, err := result.RowsAffected()
		if err != nil || rows == 0 {
			http.Error(w, "Período não encontrado", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// UpdateSupervisorProfile atualiza todas as configurações do supervisor
func UpdateSupervisorProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		// Iniciar transação
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Erro ao iniciar transação", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Atualizar valor da sessão
		sessionPrice := r.FormValue("session_price")
		_, err = tx.Exec(`
			UPDATE supervisor_profiles 
			SET session_price = NULLIF($1, '')::decimal 
			WHERE user_id = $2`,
			sessionPrice, userID)
		if err != nil {
			http.Error(w, "Erro ao atualizar valor da sessão", http.StatusInternalServerError)
			return
		}

		// Limpar horários existentes
		_, err = tx.Exec(`DELETE FROM supervisor_weekly_hours WHERE supervisor_id = $1`, userID)
		if err != nil {
			http.Error(w, "Erro ao limpar horários", http.StatusInternalServerError)
			return
		}

		// Inserir novos horários
		for day := 1; day <= 7; day++ {
			startTime := r.FormValue(fmt.Sprintf("start_time_%d", day))
			endTime := r.FormValue(fmt.Sprintf("end_time_%d", day))

			if startTime != "" && endTime != "" {
				_, err = tx.Exec(`
					INSERT INTO supervisor_weekly_hours 
					(supervisor_id, weekday, start_time, end_time)
					VALUES ($1, $2, $3, $4)`,
					userID, day, startTime, endTime)

				if err != nil {
					http.Error(w, "Erro ao salvar horários", http.StatusInternalServerError)
					return
				}
			}
		}

		// Commit da transação
		if err := tx.Commit(); err != nil {
			http.Error(w, "Erro ao finalizar alterações", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(`<div class="alert alert-success">
			<i class="fas fa-check-circle me-2"></i>
			Configurações de supervisor atualizadas com sucesso!
		</div>`))
	}
}
