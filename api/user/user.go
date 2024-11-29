// superviso/api/user/user.go
package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"superviso/api/sessions"
	"superviso/models"

	"golang.org/x/crypto/bcrypt"
)

func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			http.Error(w, "Dados inválidos", http.StatusBadRequest)
			return
		}

		// Busca o usuário pelo email
		var id int
		var hashedPassword string
		err = db.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", credentials.Email).Scan(&id, &hashedPassword)
		if err != nil {
			http.Error(w, "Usuário ou senha inválidos", http.StatusUnauthorized)
			return
		}

		// Compara a senha hasheada
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password))
		if err != nil {
			http.Error(w, "Usuário ou senha inválidos", http.StatusUnauthorized)
			return
		}

		// Cria a sessão
		session, err := sessions.GetSession(r)
		if err != nil {
			http.Error(w, "Erro ao recuperar a sessão", http.StatusInternalServerError)
			return
		}
		session.Values["user_id"] = id
		session.Save(r, w)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Login realizado com sucesso!"))
	}
}

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Erro ao decodificar dados", http.StatusBadRequest)
			return
		}

		// Hashear a senha
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Erro ao processar a senha", http.StatusInternalServerError)
			return
		}
		user.PasswordHash = string(hashedPassword)

		// Insere o usuário no banco de dados
		query := `
            INSERT INTO users (first_name, last_name, email, password_hash, crp, cpf, theory_approach, user_role, qualifications, price_per_session, sessions_availability)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        `
		_, err = db.Exec(query, user.FirstName, user.LastName, user.Email, user.PasswordHash, user.CRP, user.CPF, user.TheoryApproach, user.UserRole, user.Qualifications, user.PricePerSession, user.SessionsAvailability)
		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao salvar usuário: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Usuário registrado com sucesso!"))
	}
}
