// superviso/api/routes/routes.go
package routes

import (
	"database/sql"
	"fmt"
	"net/http"
	"superviso/api/auth"
	"superviso/api/docs"
	"superviso/api/supervisor"
	"superviso/api/user"

	"github.com/gorilla/mux"
)

func ConfigureRoutes(r *mux.Router, db *sql.DB) {
	// Arquivos estáticos para web
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

	r.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/docs.html")
	}).Methods("GET")

	// API
	r.HandleFunc("/users/register", user.Register(db)).Methods("POST")
	r.HandleFunc("/users/login", user.Login(db)).Methods("POST")
	r.HandleFunc("/users/logout", user.Logout).Methods("POST")
	r.HandleFunc("/api/docs", docs.GetDocument).Methods("GET")
	r.HandleFunc("/resend-verification", user.ResendVerification(db)).Methods("POST")

	// Rotas Páginas protegidas
	r.HandleFunc("/api/test-auth", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)
		email := r.Context().Value(auth.EmailKey).(string)
		w.Write([]byte(fmt.Sprintf("Autenticado! UserID: %d, Email: %s", userID, email)))
	})).Methods("GET")

	r.HandleFunc("/dashboard", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/dashboard.html")
	})).Methods("GET")

	r.HandleFunc("/supervisors", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/supervisors.html")
	})).Methods("GET")

	r.HandleFunc("/partials/supervisor-list", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "view/partials/supervisor_list.html")
	})).Methods("GET")

	// API protegidas
	r.HandleFunc("/profile", auth.AuthMiddleware(user.GetProfile(db))).Methods("GET")
	r.HandleFunc("/api/profile/update", auth.AuthMiddleware(user.UpdateProfile(db))).Methods("POST")
	r.HandleFunc("/api/profile/toggle-supervisor", auth.AuthMiddleware(user.ToggleSupervisor(db))).Methods("POST")
	r.HandleFunc("/api/profile/check-role", auth.AuthMiddleware(user.CheckUserRole(db))).Methods("GET")
	r.HandleFunc("/api/supervisors", auth.AuthMiddleware(supervisor.GetSupervisors(db))).Methods("GET")

	r.HandleFunc("/verify-email", user.VerifyEmail(db)).Methods("GET")
	r.HandleFunc("/resend-verification", user.ResendVerification(db)).Methods("POST")

}
