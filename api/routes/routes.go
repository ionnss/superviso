// superviso/api/routes/routes.go
package routes

import (
	"database/sql"
	"net/http"
	"superviso/api/user"

	"github.com/gorilla/mux"
)

func ConfigureRoutes(r *mux.Router, db *sql.DB) {
	// Arquivos estáticos
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("view/assets/"))))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("view/css/"))))

	// Páginas
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/index.html")
	}).Methods("GET")

	r.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/register.html")
	}).Methods("GET")

	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/login.html")
	}).Methods("GET")

	r.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/dashboard.html")
	}).Methods("GET")

	// API
	r.HandleFunc("/users/register", user.Register(db)).Methods("POST")
	r.HandleFunc("/users/login", user.Login(db)).Methods("POST")
	// r.HandleFunc("/users/logout", sessions.Logout).Methods("POST")

	// Rotas protegidas
}
