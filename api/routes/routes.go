// superviso/api/routes/routes.go
package routes

import (
	"database/sql"
	"superviso/api/user"

	"superviso/api/supervisor"

	"github.com/gorilla/mux"
)

func ConfigureRoutes(r *mux.Router, db *sql.DB) {
	// Rotas para usuários
	r.HandleFunc("/users/register", user.Register(db)).Methods("POST")
	//r.HandleFunc("/users/login", user.Login(db)).Methods("POST")

	// Rotas para supervisores
	r.HandleFunc("/supervisors", supervisor.GetSupervisors(db)).Methods("GET")
	//r.HandleFunc("/supervisors/register", supervisor.RegisterSupervisor(db)).Methods("POST")
}
