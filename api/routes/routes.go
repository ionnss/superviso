// superviso/api/routes/routes.go
package routes

import (
	"database/sql"
	"fmt"
	"net/http"
	"superviso/api/auth"
	"superviso/api/supervisor"
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

	r.HandleFunc("/dashboard", auth.AuthMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/dashboard.html")
	})).Methods("GET")

	// API
	r.HandleFunc("/users/register", user.Register(db)).Methods("POST")
	r.HandleFunc("/users/login", user.Login(db)).Methods("POST")
	r.HandleFunc("/users/logout", user.Logout).Methods("POST")

	// Rotas protegidas
	r.HandleFunc("/api/test-auth", auth.AuthMiddleware(db, func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)
		email := r.Context().Value(auth.EmailKey).(string)
		w.Write([]byte(fmt.Sprintf("Autenticado! UserID: %d, Email: %s", userID, email)))
	})).Methods("GET")

	// Rota de seleção de role
	r.HandleFunc("/select-role", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/select_role.html")
	}).Methods("GET")

	r.HandleFunc("/users/set-role", auth.AuthMiddleware(db, user.SetRole(db))).Methods("POST")

	// Rotas específicas para supervisor
	r.HandleFunc("/supervisor/schedule", auth.AuthMiddleware(db, supervisor.ConfigureSchedule(db))).Methods("GET", "POST")
}
