// superviso/api/supervisor/supervisor.go
package supervisor

import (
	"database/sql"
	"net/http"
)

func GetSupervisors(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `SELECT * FROM users WHERE user_role = 'supervisor'`
		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Erro ao buscar supervisores", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Transformar resultados em JSON e retornar
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Lista de supervisores"))
	}
}
