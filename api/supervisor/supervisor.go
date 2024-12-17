// superviso/api/supervisor/supervisor.go
package supervisor

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"superviso/models"
	"superviso/utils"
	"text/template"
	"time"
)

// Package supervisor gerencia as funcionalidades específicas de supervisores.
//
// Inclui funcionalidades para:
//   - Listagem de supervisores disponíveis
//   - Filtros por abordagem e valor
//   - Gerenciamento de horários e disponibilidade

var funcMap = template.FuncMap{
	"formatWeekday": utils.FormatWeekday,
	"formatDate":    utils.FormatDate,
	"formatTime": func(t string) string {
		if t == "" {
			return ""
		}
		// Remover os segundos se existirem
		if len(t) > 5 {
			t = t[:5]
		}
		return t
	},
}

// GetSupervisors retorna a lista de supervisores disponíveis
func GetSupervisors(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Buscar supervisores com seus horários
		rows, err := db.Query(`
				SELECT DISTINCT 
					u.id,
					u.first_name,
					u.last_name,
					u.crp,
					u.theory_approach,
					sp.session_price,
					sap.start_date,
					sap.end_date
				FROM users u
				JOIN supervisor_profiles sp ON u.id = sp.user_id
				JOIN supervisor_availability_periods sap ON sp.user_id = sap.supervisor_id
				WHERE sap.end_date >= CURRENT_DATE
				ORDER BY u.first_name, u.last_name`)
		if err != nil {
			http.Error(w, "Erro ao buscar supervisores", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var supervisors []struct {
			models.Supervisor
			StartDate   time.Time                 `json:"start_date"`
			EndDate     time.Time                 `json:"end_date"`
			WeeklyHours map[int]models.WeeklyHour `json:"weekly_hours"`
		}

		for rows.Next() {
			var s struct {
				models.Supervisor
				StartDate   time.Time
				EndDate     time.Time
				WeeklyHours map[int]models.WeeklyHour
			}
			s.WeeklyHours = make(map[int]models.WeeklyHour)

			err := rows.Scan(
				&s.UserID, &s.FirstName, &s.LastName,
				&s.CRP, &s.TheoryApproach, &s.SessionPrice,
				&s.StartDate, &s.EndDate)
			if err != nil {
				http.Error(w, "Erro ao ler dados", http.StatusInternalServerError)
				return
			}

			// Buscar horários do supervisor
			hourRows, err := db.Query(`
				SELECT 
					weekday,
					TO_CHAR(start_time, 'HH24:MI') as start_time,
					TO_CHAR(end_time, 'HH24:MI') as end_time
				FROM supervisor_weekly_hours 
				WHERE supervisor_id = $1 
				ORDER BY weekday`, s.UserID)
			if err != nil {
				http.Error(w, "Erro ao buscar horários", http.StatusInternalServerError)
				return
			}
			defer hourRows.Close()

			for hourRows.Next() {
				var h models.WeeklyHour
				if err := hourRows.Scan(&h.Weekday, &h.StartTime, &h.EndTime); err != nil {
					http.Error(w, "Erro ao ler horários", http.StatusInternalServerError)
					return
				}
				s.WeeklyHours[h.Weekday] = h
			}

			supervisors = append(supervisors, struct {
				models.Supervisor
				StartDate   time.Time                 "json:\"start_date\""
				EndDate     time.Time                 "json:\"end_date\""
				WeeklyHours map[int]models.WeeklyHour "json:\"weekly_hours\""
			}(s))
		}

		// Se for requisição HTMX, retorna HTML
		if r.Header.Get("HX-Request") == "true" {
			tmpl := template.Must(template.New("supervisor_list.html").
				Funcs(funcMap).
				ParseFiles("view/partials/supervisor_list.html"))
			tmpl.Execute(w, supervisors)
			return
		}

		// Senão retorna JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(supervisors)
	}
}
