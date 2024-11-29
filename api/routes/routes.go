// superviso/api/routes/routes.go
package routes

import (
	"database/sql"
	"superviso/api/sessions"
	"superviso/api/supervisor"
	"superviso/api/user"

	"github.com/gorilla/mux"
)

func ConfigureRoutes(r *mux.Router, db *sql.DB) {
	// Rotas públicas
	r.HandleFunc("/users/register", user.Register(db)).Methods("POST")
	r.HandleFunc("/users/login", user.Login(db)).Methods("POST")

	// Rotas protegidas
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(sessions.AuthMiddleware)
	protected.HandleFunc("/supervisors", supervisor.GetSupervisors(db)).Methods("GET")
}
