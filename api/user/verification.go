package user

import (
	"database/sql"
	"fmt"
	"net/http"
	"superviso/api/auth"
	"superviso/api/email"
	"time"
)

func SendVerificationEmail(db *sql.DB, userID int, userEmail string) error {
	// Gerar token e data de expiração
	token, err := auth.GenerateVerificationToken()
	if err != nil {
		return fmt.Errorf("erro ao gerar token: %v", err)
	}

	expiry := auth.GenerateVerificationTokenExpiry()

	// Atualizar usuário com token
	_, err = db.Exec(`
		UPDATE users 
		SET verification_token = $1, 
		    verification_token_expires = $2 
		WHERE id = $3`,
		token, expiry, userID)
	if err != nil {
		return fmt.Errorf("erro ao salvar token: %v", err)
	}

	// Enviar email
	if err := email.SendVerificationEmail(userEmail, token); err != nil {
		return fmt.Errorf("erro ao enviar email: %v", err)
	}

	return nil
}

func VerifyEmail(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "Token inválido", http.StatusBadRequest)
			return
		}

		var userID int
		var tokenExpiry time.Time
		err := db.QueryRow(`
			SELECT id, verification_token_expires 
			FROM users 
			WHERE verification_token = $1 
			AND email_verified = false`,
			token).Scan(&userID, &tokenExpiry)

		if err == sql.ErrNoRows {
			http.Error(w, "Token inválido ou já utilizado", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "Erro ao verificar token", http.StatusInternalServerError)
			return
		}

		if time.Now().After(tokenExpiry) {
			http.Error(w, "Token expirado", http.StatusBadRequest)
			return
		}

		// Atualizar usuário como verificado
		_, err = db.Exec(`
			UPDATE users 
			SET email_verified = true, 
			    verification_token = null, 
			    verification_token_expires = null 
			WHERE id = $1`,
			userID)
		if err != nil {
			http.Error(w, "Erro ao verificar email", http.StatusInternalServerError)
			return
		}

		// Redirecionar para página de login com mensagem de sucesso
		http.Redirect(w, r, "/login?msg=email_verified", http.StatusSeeOther)
	}
}

func ResendVerification(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		if email == "" {
			http.Error(w, "Email não fornecido", http.StatusBadRequest)
			return
		}

		var userID int
		err := db.QueryRow("SELECT id FROM users WHERE email = $1 AND email_verified = false", email).Scan(&userID)
		if err == sql.ErrNoRows {
			http.Error(w, "Email não encontrado ou já verificado", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "Erro ao buscar usuário", http.StatusInternalServerError)
			return
		}

		if err := SendVerificationEmail(db, userID, email); err != nil {
			http.Error(w, "Erro ao reenviar email", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(`<div class="alert alert-success">Email de verificação reenviado!</div>`))
	}
}
