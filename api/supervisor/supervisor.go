// superviso/api/supervisor/supervisor.go
package supervisor

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"

	"superviso/models"
)

func GetSupervisors(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parâmetros de filtro
		approach := r.URL.Query().Get("approach")
		maxPrice := r.URL.Query().Get("max_price")

		query := `
			SELECT 
				u.id,
				u.first_name,
				u.last_name,
				u.crp,
				u.theory_approach,
				sp.session_price,
				sp.available_days,
				sp.start_time,
				sp.end_time,
				sp.created_at
			FROM users u
			INNER JOIN supervisor_profiles sp ON u.id = sp.user_id
			WHERE 1=1
		`
		var params []interface{}

		if approach != "" {
			query += " AND u.theory_approach ILIKE $1"
			params = append(params, "%"+approach+"%")
		}

		if maxPrice != "" {
			query += " AND sp.session_price <= $2"
			params = append(params, maxPrice)
		}

		query += " ORDER BY sp.created_at DESC"

		// Executa a query e obtém os resultados
		rows, err := db.Query(query, params...)
		if err != nil {
			http.Error(w, "Erro ao buscar supervisores", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var supervisors []models.Supervisor
		for rows.Next() {
			var s models.Supervisor
			err := rows.Scan(
				&s.UserID, &s.FirstName, &s.LastName, &s.CRP,
				&s.TheoryApproach, &s.SessionPrice, &s.AvailableDays,
				&s.StartTime, &s.EndTime, &s.CreatedAt,
			)
			if err != nil {
				http.Error(w, "Erro ao ler dados", http.StatusInternalServerError)
				return
			}
			supervisors = append(supervisors, s)
		}

		// Se a requisição for HTMX, retorna HTML
		if r.Header.Get("HX-Request") == "true" {
			tmpl := template.Must(template.ParseFiles("view/partials/supervisor_list.html"))
			tmpl.Execute(w, supervisors)
			return
		}

		// Se não for HTMX, retorna JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(supervisors)
	}
}
