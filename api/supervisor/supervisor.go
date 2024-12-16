// superviso/api/supervisor/supervisor.go
package supervisor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"superviso/models"
)

// Package supervisor gerencia as funcionalidades específicas de supervisores.
//
// Inclui funcionalidades para:
//   - Listagem de supervisores disponíveis
//   - Filtros por abordagem e valor
//   - Gerenciamento de horários e disponibilidade
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
				sp.created_at
			FROM users u
			INNER JOIN supervisor_profiles sp ON u.id = sp.user_id
			WHERE 1=1
		`
		var params []interface{}
		paramCount := 1

		if approach != "" {
			query += fmt.Sprintf(" AND u.theory_approach ILIKE $%d", paramCount)
			params = append(params, "%"+approach+"%")
			paramCount++
		}

		if maxPrice != "" {
			query += fmt.Sprintf(" AND sp.session_price <= $%d", paramCount)
			price, err := strconv.ParseFloat(maxPrice, 64)
			if err != nil {
				http.Error(w, "Valor máximo inválido", http.StatusBadRequest)
				return
			}
			params = append(params, price)
			paramCount++
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
				&s.TheoryApproach, &s.SessionPrice, &s.CreatedAt,
			)
			if err != nil {
				http.Error(w, "Erro ao ler dados", http.StatusInternalServerError)
				return
			}
			supervisors = append(supervisors, s)
		}

		// Se a requisição for HTMX, retorna HTML
		if r.Header.Get("HX-Request") == "true" {
			funcMap := template.FuncMap{
				"formatTime": func(t string) string {
					timeObj, err := time.Parse(time.RFC3339, t)
					if err != nil {
						// Tentar outro formato caso o primeiro falhe
						timeObj, err = time.Parse("2006-01-02T15:04:05Z", t)
						if err != nil {
							return t
						}
					}
					return timeObj.Format("15:04")
				},
				"formatDays": func(days string) string {
					dayMap := map[string]string{
						"1": "Segunda",
						"2": "Terça",
						"3": "Quarta",
						"4": "Quinta",
						"5": "Sexta",
						"6": "Sábado",
						"7": "Domingo",
					}

					var result []string
					for _, day := range strings.Split(days, ",") {
						if name, ok := dayMap[day]; ok {
							result = append(result, name)
						}
					}
					return strings.Join(result, ", ")
				},
			}

			tmpl := template.Must(template.New("supervisor_list.html").
				Funcs(funcMap).
				ParseFiles("view/partials/supervisor_list.html"))
			tmpl.Execute(w, supervisors)
			return
		}

		// Se não for HTMX, retorna JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(supervisors)
	}
}
