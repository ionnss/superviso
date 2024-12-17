package supervisor

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"superviso/api/auth"
)

// ToggleDayHours é chamado via HTMX quando um dia é selecionado/desmarcado
func ToggleDayHours(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)
		dayIndex := r.FormValue("value")
		isChecked := r.FormValue("checked") == "true"

		if isChecked {
			// Buscar horários existentes para este dia
			var startTime, endTime string
			err := db.QueryRow(`
                SELECT start_time, end_time 
                FROM supervisor_weekly_hours 
                WHERE supervisor_id = $1 AND weekday = $2`,
				userID, dayIndex).Scan(&startTime, &endTime)

			if err != nil && err != sql.ErrNoRows {
				http.Error(w, "Erro ao buscar horários", http.StatusInternalServerError)
				return
			}

			// Retorna os campos de horário
			tmpl := template.Must(template.New("hours").Parse(`
                <div class="row">
                    <div class="col-md-5">
                        <label class="form-label">Início</label>
                        <input type="time" class="form-control" 
                               name="start_time_{{.Day}}" 
                               value="{{.Start}}">
                    </div>
                    <div class="col-md-5">
                        <label class="form-label">Fim</label>
                        <input type="time" class="form-control" 
                               name="end_time_{{.Day}}" 
                               value="{{.End}}">
                    </div>
                </div>
            `))

			tmpl.Execute(w, map[string]interface{}{
				"Day":   dayIndex,
				"Start": startTime,
				"End":   endTime,
			})
		}
	}
}

// UpdateAvailability atualiza todas as configurações do supervisor
func UpdateAvailability(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Erro ao iniciar transação", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// 1. Atualizar valor da sessão
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

		// 2. Atualizar período de disponibilidade
		startDate := r.FormValue("availability_start")
		endDate := r.FormValue("availability_end")
		_, err = tx.Exec(`
            INSERT INTO supervisor_availability_periods 
            (supervisor_id, start_date, end_date)
            VALUES ($1, $2::date, $3::date)
            ON CONFLICT ON CONSTRAINT unique_supervisor_period 
            DO UPDATE SET 
                start_date = EXCLUDED.start_date,
                end_date = EXCLUDED.end_date`,
			userID, startDate, endDate)
		if err != nil {
			http.Error(w, "Erro ao atualizar período", http.StatusInternalServerError)
			return
		}

		// 3. Limpar horários antigos
		_, err = tx.Exec(`DELETE FROM supervisor_weekly_hours WHERE supervisor_id = $1`, userID)
		if err != nil {
			http.Error(w, "Erro ao limpar horários", http.StatusInternalServerError)
			return
		}

		// 4. Inserir novos horários
		for day := 1; day <= 7; day++ {
			if r.FormValue(fmt.Sprintf("day_%d", day)) == "on" {
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
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Erro ao finalizar alterações", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(`
            <div class="alert alert-success">
                <i class="fas fa-check-circle me-2"></i>
                Configurações atualizadas com sucesso!
            </div>
        `))
	}
}

// GetSupervisorAvailability busca todos os dados de disponibilidade do supervisor
func GetSupervisorAvailability(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		// 1. Buscar valor da sessão
		var sessionPrice float64
		err := db.QueryRow(`
			SELECT session_price 
			FROM supervisor_profiles 
			WHERE user_id = $1`,
			userID).Scan(&sessionPrice)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "Erro ao buscar valor da sessão", http.StatusInternalServerError)
			return
		}

		// 2. Buscar período de disponibilidade
		var startDate, endDate string
		err = db.QueryRow(`
			SELECT start_date, end_date 
			FROM supervisor_availability_periods 
			WHERE supervisor_id = $1 
			AND end_date >= CURRENT_DATE
			ORDER BY start_date LIMIT 1`,
			userID).Scan(&startDate, &endDate)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "Erro ao buscar período", http.StatusInternalServerError)
			return
		}

		// 3. Buscar horários semanais
		rows, err := db.Query(`
			SELECT weekday, start_time, end_time 
			FROM supervisor_weekly_hours 
			WHERE supervisor_id = $1 
			ORDER BY weekday`,
			userID)
		if err != nil {
			http.Error(w, "Erro ao buscar horários", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		weeklyHours := make(map[int]struct {
			StartTime string
			EndTime   string
		})

		for rows.Next() {
			var day int
			var start, end string
			if err := rows.Scan(&day, &start, &end); err != nil {
				http.Error(w, "Erro ao ler horários", http.StatusInternalServerError)
				return
			}
			weeklyHours[day] = struct {
				StartTime string
				EndTime   string
			}{start, end}
		}

		// Preparar dados para o template
		data := struct {
			SessionPrice float64
			StartDate    string
			EndDate      string
			WeeklyHours  map[int]struct{ StartTime, EndTime string }
			WeekDays     []int
		}{
			SessionPrice: sessionPrice,
			StartDate:    startDate,
			EndDate:      endDate,
			WeeklyHours:  weeklyHours,
			WeekDays:     []int{1, 2, 3, 4, 5, 6, 7},
		}

		// Renderizar template
		tmpl := template.Must(template.ParseFiles("view/partials/supervisor_fields.html"))
		tmpl.Execute(w, data)
	}
}
