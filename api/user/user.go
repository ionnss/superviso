package user

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"superviso/api/sessions"
	"superviso/models"

	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLength = 8
	MaxLoginAttempts  = 5
)

// sendHTMLResponse envia uma resposta HTML com uma mensagem estilizada
func sendHTMLResponse(w http.ResponseWriter, status int, message string, isError bool) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)

	alertClass := "alert-success"
	if isError {
		alertClass = "alert-danger"
	}

	html := fmt.Sprintf(`
		<div class="container mt-3">
			<div class="alert %s alert-dismissible fade show" role="alert">
				%s
				<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
			</div>
		</div>
	`, alertClass, message)

	fmt.Fprint(w, html)
}

// Register handles user registration for both supervisor and supervisionated
func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			log.Printf("Erro ao processar formulário: %v", err)
			http.Error(w, "Erro ao processar os dados do formulário.", http.StatusBadRequest)
			return
		}

		// Log dos dados recebidos
		log.Printf("URL Path: %s", r.URL.Path)
		log.Printf("Form values: %+v", r.Form)
		log.Printf("PostForm values: %+v", r.PostForm)

		// Determina o tipo de usuário baseado na URL
		userType := "supervisionated" // valor padrão
		if r.URL.Path == "/users/register/supervisor" {
			userType = "supervisor"
		}

		// Coleta os dados do formulário
		formData := map[string]string{
			"first_name":      r.PostForm.Get("first_name"),
			"last_name":       r.PostForm.Get("last_name"),
			"email":           r.PostForm.Get("email"),
			"password":        r.PostForm.Get("password"),
			"cpf":             r.PostForm.Get("cpf"),
			"crp":             r.PostForm.Get("crp"),
			"theory_approach": r.PostForm.Get("theory_approach"),
			"qualifications":  r.PostForm.Get("qualifications"),
		}

		// Log dos dados processados
		log.Printf("Dados processados: %+v", formData)

		// Validação de campos obrigatórios
		for field, value := range formData {
			if value == "" {
				log.Printf("Campo obrigatório faltando: %s", field)
				http.Error(w, "Todos os campos são obrigatórios.", http.StatusBadRequest)
				return
			}
		}

		// Gera o hash da senha
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(formData["password"]), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Erro ao gerar hash da senha: %v", err)
			http.Error(w, "Erro ao processar a senha.", http.StatusInternalServerError)
			return
		}

		// Construir query dinamicamente
		query := `
			INSERT INTO ` + userType + ` (
				first_name, last_name, email, password_hash, 
				cpf, crp, theory_approach, qualifications, user_role
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9
			) RETURNING id`

		var userID int
		err = db.QueryRow(
			query,
			formData["first_name"],
			formData["last_name"],
			formData["email"],
			string(hashedPassword),
			formData["cpf"],
			formData["crp"],
			formData["theory_approach"],
			formData["qualifications"],
			userType,
		).Scan(&userID)

		if err != nil {
			log.Printf("Erro ao inserir usuário no banco: %v", err)
			http.Error(w, "Erro ao registrar usuário no banco de dados.", http.StatusInternalServerError)
			return
		}

		log.Printf("Usuário registrado com sucesso. ID: %d, Type: %s", userID, userType)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

// LoginHandler handles the login process
func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Printf("Erro ao processar formulário: %v", err)
			sendHTMLResponse(w, http.StatusBadRequest, "Erro ao processar os dados do formulário.", true)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "" || password == "" {
			sendHTMLResponse(w, http.StatusBadRequest, "Email e senha são obrigatórios.", true)
			return
		}

		// Primeiro tenta encontrar na tabela de supervisores
		var supervisor models.Supervisor
		err = db.QueryRow(`
			SELECT id, email, password_hash, user_role, failed_login_attempts, last_failed_login 
			FROM supervisor WHERE email = $1`,
			email).Scan(
			&supervisor.ID, &supervisor.Email, &supervisor.PasswordHash,
			&supervisor.UserRole, &supervisor.FailedLoginAttempts, &supervisor.LastFailedLogin)

		var user models.AuthUser
		if err == sql.ErrNoRows {
			// Se não encontrou supervisor, tenta na tabela de supervisionados
			var supervisionated models.Supervisionated
			err = db.QueryRow(`
				SELECT id, email, password_hash, user_role, failed_login_attempts, last_failed_login 
				FROM supervisionated WHERE email = $1`,
				email).Scan(
				&supervisionated.ID, &supervisionated.Email, &supervisionated.PasswordHash,
				&supervisionated.UserRole, &supervisionated.FailedLoginAttempts, &supervisionated.LastFailedLogin)

			if err == sql.ErrNoRows {
				sendHTMLResponse(w, http.StatusUnauthorized, "Usuário ou senha inválidos.", true)
				return
			} else if err != nil {
				log.Printf("Erro ao buscar supervisionado: %v", err)
				sendHTMLResponse(w, http.StatusInternalServerError, "Erro ao processar login.", true)
				return
			}
			user = &supervisionated
		} else if err != nil {
			log.Printf("Erro ao buscar supervisor: %v", err)
			sendHTMLResponse(w, http.StatusInternalServerError, "Erro ao processar login.", true)
			return
		} else {
			user = &supervisor
		}

		// Verifica se a conta está bloqueada
		if user.GetFailedLoginAttempts() >= 5 {
			lastAttempt := user.GetLastFailedLogin()
			if time.Since(lastAttempt) < 15*time.Minute {
				sendHTMLResponse(w, http.StatusTooManyRequests, "Conta bloqueada. Tente novamente mais tarde.", true)
				return
			}
		}

		// Verifica a senha
		if err := bcrypt.CompareHashAndPassword(
			[]byte(user.GetPasswordHash()),
			[]byte(password)); err != nil {

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

			sendHTMLResponse(w, http.StatusUnauthorized, "Usuário ou senha inválidos.", true)
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
			log.Printf("Erro ao criar sessão: %v", err)
			sendHTMLResponse(w, http.StatusInternalServerError, "Erro ao criar sessão.", true)
			return
		}

		// Armazena informações do usuário na sessão
		session.Values["user_id"] = user.GetID()
		session.Values["user_role"] = user.GetUserRole()
		session.Values["email"] = user.GetEmail()

		if err := session.Save(r, w); err != nil {
			log.Printf("Erro ao salvar sessão: %v", err)
			sendHTMLResponse(w, http.StatusInternalServerError, "Erro ao salvar sessão.", true)
			return
		}

		sendHTMLResponse(w, http.StatusOK, "Login realizado com sucesso!", false)
	}
}
