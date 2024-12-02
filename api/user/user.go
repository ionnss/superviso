package user

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"superviso/api/sessions"
	"superviso/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength = 8
	MaxLoginAttempts  = 5
)

// sendResponse é uma função auxiliar para enviar respostas JSON padronizadas
func sendResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

// Register handles user registration for both supervisor and supervisionated
func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			sendResponse(w, http.StatusBadRequest, "Erro ao processar os dados do formulário.")
			return
		}

		userType := r.FormValue("user_role")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			sendResponse(w, http.StatusInternalServerError, "Erro ao processar a senha.")
			return
		}

		// Dados comuns para ambos os tipos
		userData := map[string]interface{}{
			"first_name":      r.FormValue("first_name"),
			"last_name":       r.FormValue("last_name"),
			"email":           r.FormValue("email"),
			"password_hash":   string(hashedPassword),
			"cpf":             r.FormValue("cpf"),
			"crp":             r.FormValue("crp"),
			"theory_approach": r.FormValue("theory_approach"),
			"qualifications":  r.FormValue("qualifications"),
			"user_role":       userType,
		}

		var table string
		if userType == "supervisor" {
			table = "supervisor"
		} else {
			table = "supervisionated"
		}

		// Construir query dinamicamente
		query := `
			INSERT INTO ` + table + ` (
				first_name, last_name, email, password_hash, 
				cpf, crp, theory_approach, qualifications, user_role
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9
			) RETURNING id`

		var userID int
		err = db.QueryRow(
			query,
			userData["first_name"],
			userData["last_name"],
			userData["email"],
			userData["password_hash"],
			userData["cpf"],
			userData["crp"],
			userData["theory_approach"],
			userData["qualifications"],
			userData["user_role"],
		).Scan(&userID)

		if err != nil {
			log.Printf("Erro ao registrar usuário: %v", err)
			sendResponse(w, http.StatusInternalServerError, "Erro ao registrar usuário.")
			return
		}

		sendResponse(w, http.StatusCreated, "Usuário registrado com sucesso!")
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var credentials models.LoginCredentials
		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			sendResponse(w, http.StatusBadRequest, "Dados inválidos.")
			return
		}

		// Primeiro tenta encontrar na tabela de supervisores
		var supervisor models.Supervisor
		err := db.QueryRow(`
			SELECT id, email, password_hash, user_role, failed_login_attempts, last_failed_login 
			FROM supervisor WHERE email = $1`,
			credentials.Email).Scan(
			&supervisor.ID, &supervisor.Email, &supervisor.PasswordHash,
			&supervisor.UserRole, &supervisor.FailedLoginAttempts, &supervisor.LastFailedLogin)

		var user models.AuthUser
		if err == sql.ErrNoRows {
			// Se não encontrou supervisor, tenta na tabela de supervisionados
			var supervisionated models.Supervisionated
			err = db.QueryRow(`
				SELECT id, email, password_hash, user_role, failed_login_attempts, last_failed_login 
				FROM supervisionated WHERE email = $1`,
				credentials.Email).Scan(
				&supervisionated.ID, &supervisionated.Email, &supervisionated.PasswordHash,
				&supervisionated.UserRole, &supervisionated.FailedLoginAttempts, &supervisionated.LastFailedLogin)

			if err == sql.ErrNoRows {
				sendResponse(w, http.StatusUnauthorized, "Usuário ou senha inválidos.")
				return
			} else if err != nil {
				sendResponse(w, http.StatusInternalServerError, "Erro ao processar login.")
				return
			}
			user = &supervisionated
		} else if err != nil {
			sendResponse(w, http.StatusInternalServerError, "Erro ao processar login.")
			return
		} else {
			user = &supervisor
		}

		// Verifica se a conta está bloqueada
		if user.GetFailedLoginAttempts() >= 5 {
			lastAttempt := user.GetLastFailedLogin()
			if time.Since(lastAttempt) < 15*time.Minute {
				sendResponse(w, http.StatusTooManyRequests, "Conta bloqueada. Tente novamente mais tarde.")
				return
			}
		}

		// Verifica a senha
		if err := bcrypt.CompareHashAndPassword(
			[]byte(user.GetPasswordHash()),
			[]byte(credentials.Password)); err != nil {

			// Atualiza contagem de tentativas falhas
			var table string
			if user.GetUserRole() == "supervisor" {
				table = "supervisor"
			} else {
				table = "supervisionated"
			}

			_, err := db.Exec(`
				UPDATE `+table+` 
				SET failed_login_attempts = failed_login_attempts + 1,
					last_failed_login = CURRENT_TIMESTAMP
				WHERE id = $1`, user.GetID())

			if err != nil {
				log.Printf("Erro ao atualizar tentativas de login: %v", err)
			}

			sendResponse(w, http.StatusUnauthorized, "Usuário ou senha inválidos.")
			return
		}

		// Reset failed attempts on successful login
		var table string
		if user.GetUserRole() == "supervisor" {
			table = "supervisor"
		} else {
			table = "supervisionated"
		}

		_, err = db.Exec(`
			UPDATE `+table+` 
			SET failed_login_attempts = 0,
				last_failed_login = NULL
			WHERE id = $1`, user.GetID())

		if err != nil {
			log.Printf("Erro ao resetar tentativas de login: %v", err)
		}

		// Cria sessão
		session, err := sessions.GetSession(r)
		if err != nil {
			sendResponse(w, http.StatusInternalServerError, "Erro ao criar sessão.")
			return
		}

		// Armazena informações do usuário na sessão
		session.Values["user_id"] = user.GetID()
		session.Values["user_role"] = user.GetUserRole()
		session.Values["email"] = user.GetEmail()

		if err := session.Save(r, w); err != nil {
			sendResponse(w, http.StatusInternalServerError, "Erro ao salvar sessão.")
			return
		}

		sendResponse(w, http.StatusOK, "Login realizado com sucesso!")
	}
}
