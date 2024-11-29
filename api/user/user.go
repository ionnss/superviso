package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"superviso/api/sessions"
	"superviso/models"

	"golang.org/x/crypto/bcrypt"
)

// Login realiza o login do usuário, verifica credenciais e cria uma sessão.
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

		sendResponse(w, http.StatusOK, "Login realizado com sucesso!")
	}
}

// Register realiza o registro de um novo usuário.
func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Processar os dados do formulário
		err := r.ParseForm()
		if err != nil {
			sendResponse(w, http.StatusBadRequest, "Erro ao processar o formulário.")
			return
		}

		// Capturar os campos do formulário
		user := models.User{
			FirstName:       r.FormValue("firstname"),
			LastName:        r.FormValue("lastname"),
			Email:           r.FormValue("email"),
			Password:        r.FormValue("password"),
			CPF:             r.FormValue("cpf"),
			CRP:             r.FormValue("crp"),
			TheoryApproach:  r.FormValue("approach"),
			Qualifications:  r.FormValue("qualifications"),
			UserRole:        r.FormValue("usertype"),
			PricePerSession: parseFloat(r.FormValue("price_per_session")),
		}

		// Validações de entrada
		if err := validateUser(&user); err != nil {
			sendResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Hashear a senha
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			sendResponse(w, http.StatusInternalServerError, "Erro ao processar a senha.")
			return
		}
		user.PasswordHash = string(hashedPassword)

		// Verificar se o e-mail já está registrado
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", user.Email).Scan(&exists)
		if err != nil {
			sendResponse(w, http.StatusInternalServerError, "Erro ao verificar e-mail.")
			return
		}
		if exists {
			sendResponse(w, http.StatusBadRequest, "E-mail já registrado.")
			return
		}

		// Inserir o usuário no banco de dados
		var userID int
		query := `
			INSERT INTO users (first_name, last_name, email, password_hash, crp, cpf, theory_approach, user_role, qualifications, price_per_session, sessions_availability)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id
		`
		err = db.QueryRow(query, user.FirstName, user.LastName, user.Email, user.PasswordHash, user.CRP, user.CPF, user.TheoryApproach, user.UserRole, user.Qualifications, user.PricePerSession, user.SessionsAvailability).Scan(&userID)
		if err != nil {
			sendResponse(w, http.StatusInternalServerError, fmt.Sprintf("Erro ao salvar usuário: %v", err))
			return
		}

		// Processar horários
		daysOfWeek := []string{"segunda", "terça", "quarta", "quinta", "sexta", "sábado", "domingo"}
		for _, day := range daysOfWeek {
			hour := r.FormValue(fmt.Sprintf("availability[%s]", day))
			if hour != "" { // Apenas salva se o horário foi preenchido
				query := "INSERT INTO user_availability (user_id, day, time) VALUES ($1, $2, $3)"
				_, err := db.Exec(query, userID, day, hour)
				if err != nil {
					sendResponse(w, http.StatusInternalServerError, "Erro ao salvar disponibilidade.")
					return
				}
			}
		}

		sendResponse(w, http.StatusCreated, "Usuário registrado com sucesso!")
	}
}

func parseFloat(value string) float64 {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0
	}
	return f
}

// Funções auxiliares

// sendResponse envia uma resposta JSON ao cliente.
func sendResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

// validateUser valida os campos do usuário antes de salvar.
func validateUser(user *models.User) error {
	if !isValidEmail(user.Email) {
		return fmt.Errorf("email inválido")
	}
	if user.CPF != "" && !isValidCPF(user.CPF) {
		return fmt.Errorf("CPF inválido")
	}
	if user.CRP != "" && !isValidCRP(user.CRP) {
		return fmt.Errorf("CRP inválido")
	}
	return nil
}

// isValidEmail valida o formato do e-mail.
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

// isValidCPF valida o formato e os dígitos do CPF.
func isValidCPF(cpf string) bool {
	cpf = regexp.MustCompile(`[^\d]`).ReplaceAllString(cpf, "")

	if len(cpf) != 11 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		digit := int(cpf[i] - '0')
		sum += digit * (10 - i)
	}

	firstCheck := (sum * 10) % 11
	if firstCheck == 10 {
		firstCheck = 0
	}
	if firstCheck != int(cpf[9]-'0') {
		return false
	}

	sum = 0
	for i := 0; i < 10; i++ {
		digit := int(cpf[i] - '0')
		sum += digit * (11 - i)
	}

	secondCheck := (sum * 10) % 11
	if secondCheck == 10 {
		secondCheck = 0
	}
	return secondCheck == int(cpf[10]-'0')
}

// isValidCRP valida o formato do CRP (ex.: SP-12345).
func isValidCRP(crp string) bool {
	re := regexp.MustCompile(`^[A-Z]{2}-\d{4,5}$`)
	return re.MatchString(crp)
}
