// superviso/api/supervisor/supervisor.go
package supervisor

import (
	"database/sql"
	"net/http"
)

func ConfigureSchedule(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			http.ServeFile(w, r, "view/supervisor/schedule.html")
			return
		}

		// Lógica para POST será implementada depois
		http.Error(w, "Método não implementado", http.StatusNotImplemented)
	}
}
