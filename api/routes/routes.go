// superviso/api/routes/routes.go
package routes

import (
	"database/sql"
	"net/http"
	"superviso/api/sessions"
	"superviso/api/supervisor"
	"superviso/api/user"

	"github.com/gorilla/mux"
)

func ConfigureRoutes(r *mux.Router, db *sql.DB) {
	// Servir arquivos estáticos em /assets/
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("view/assets/"))))

	// Servir arquivos estáticos em /css/
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("view/css/"))))

	// Rota raiz para servir o arquivo HTML
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/index.html")
	}).Methods("GET")

	// Rotas públicas
	r.HandleFunc("/users/register", user.Register(db)).Methods("POST")
	r.HandleFunc("/users/login", user.Login(db)).Methods("POST")
	r.HandleFunc("/users/logout", sessions.Logout).Methods("POST")

	// Rotas protegidas
	protected := r.PathPrefix("/").Subrouter()
	protected.Use(sessions.AuthMiddleware)
	protected.HandleFunc("/supervisors", supervisor.GetSupervisors(db)).Methods("GET")
}
