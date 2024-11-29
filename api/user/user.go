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

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Processar os dados do formulário
		err := r.ParseForm()
		if err != nil {
			sendResponse(w, http.StatusBadRequest, "Erro ao processar os dados do formulário.")
			return
		}

		// Capturar os campos do formulário
		user := models.User{
			FirstName:      r.FormValue("firstname"),
			LastName:       r.FormValue("lastname"),
			Email:          r.FormValue("email"),
			Password:       r.FormValue("password"),
			CPF:            r.FormValue("cpf"),
			CRP:            r.FormValue("crp"),
			TheoryApproach: r.FormValue("approach"),
			Qualifications: r.FormValue("qualifications"),
			UserRole:       r.FormValue("usertype"),
		}

		// Validação de entrada
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

		// Verificar duplicidade de email, CPF e CRP
		if err := checkDuplicateFields(db, user); err != nil {
			sendResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// Inserir o usuário na tabela `users`
		var userID int
		query := `
			INSERT INTO users (first_name, last_name, email, password_hash, cpf, crp, theory_approach, qualifications, user_role)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id
		`
		err = db.QueryRow(query, user.FirstName, user.LastName, user.Email, user.PasswordHash, user.CPF, user.CRP, user.TheoryApproach, user.Qualifications, user.UserRole).Scan(&userID)
		if err != nil {
			sendResponse(w, http.StatusInternalServerError, fmt.Sprintf("Erro ao salvar usuário: %v", err))
			return
		}

		// Inserir na tabela `supervisor_availability` caso seja supervisor
		if user.UserRole == "supervisor" {
			price, _ := strconv.ParseFloat(r.FormValue("price_per_session"), 64)
			daysOfWeek := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
			for _, day := range daysOfWeek {
				startTime := r.FormValue(fmt.Sprintf("availability[%s][start]", day))
				endTime := r.FormValue(fmt.Sprintf("availability[%s][end]", day))
				if startTime != "" && endTime != "" {
					query := `
						INSERT INTO supervisor_availability (user_id, availability_day, availability_time, price_per_session)
						VALUES ($1, $2, $3, $4)
					`
					_, err = db.Exec(query, userID, day, startTime, price)
					if err != nil {
						sendResponse(w, http.StatusInternalServerError, "Erro ao salvar disponibilidade.")
						return
					}
				}
			}
		}

		sendResponse(w, http.StatusCreated, "Usuário registrado com sucesso!")
	}
}

// Verifica duplicidade de email, CPF e CRP
func checkDuplicateFields(db *sql.DB, user models.User) error {
	fields := []struct {
		query string
		value string
		name  string
	}{
		{"SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", user.Email, "E-mail"},
		{"SELECT EXISTS(SELECT 1 FROM users WHERE cpf = $1)", user.CPF, "CPF"},
		{"SELECT EXISTS(SELECT 1 FROM users WHERE crp = $1)", user.CRP, "CRP"},
	}
	for _, field := range fields {
		var exists bool
		err := db.QueryRow(field.query, field.value).Scan(&exists)
		if err != nil {
			return fmt.Errorf("erro ao verificar duplicidade de %s", field.name)
		}
		if exists {
			return fmt.Errorf("%s já registrado", field.name)
		}
	}
	return nil
}

// Validação de campos
func validateUser(user *models.User) error {
	if !isValidEmail(user.Email) {
		return fmt.Errorf("email inválido")
	}
	if !isValidCPF(user.CPF) {
		return fmt.Errorf("CPF inválido")
	}
	return nil
}

// Validações adicionais
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func isValidCPF(cpf string) bool {
	// Adicione a lógica de validação do CPF aqui
	return len(cpf) == 11
}

// Envia uma resposta JSON
func sendResponse(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func GetRoleFields() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obtenha o valor da opção selecionada
		role := r.URL.Query().Get("usertype")

		var html string

		if role == "supervisor" {
			// Formulário para Supervisor
			html = `
            <div class="container mt-4" style="max-width: 600px;">
				<h3 class="text-center mb-4">Defina o valor da sessão e disponibilidade de horários</h3>
				<!-- Campo para preço da sessão -->
				<div class="mb-4">
					<label for="price" class="form-label">Preço por Sessão (R$)</label>
					<input type="number" class="form-control" id="price" name="price_per_session" step="0.01" placeholder="Ex: 150.00">
				</div>
				<!-- Disponibilidade por dia -->
				<div class="mb-3">
					<label for="monday" class="form-label">Segunda-Feira</label>
					<div class="row g-2">
						<div class="col">
							<label class="form-label small">Das</label>
							<input type="time" class="form-control" name="availability[monday][start]" required>
						</div>
						<div class="col">
							<label class="form-label small">Até</label>
							<input type="time" class="form-control" name="availability[monday][end]" required>
						</div>
					</div>
				</div>
				<div class="mb-3">
					<label for="tuesday" class="form-label">Terça-Feira</label>
					<div class="row g-2">
						<div class="col">
							<label class="form-label small">Das</label>
							<input type="time" class="form-control" name="availability[tuesday][start]" required>
						</div>
						<div class="col">
							<label class="form-label small">Até</label>
							<input type="time" class="form-control" name="availability[tuesday][end]" required>
						</div>
					</div>
				</div>
				<div class="mb-3">
					<label for="wednesday" class="form-label">Quarta-Feira</label>
					<div class="row g-2">
						<div class="col">
							<label class="form-label small">Das</label>
							<input type="time" class="form-control" name="availability[wednesday][start]" required>
						</div>
						<div class="col">
							<label class="form-label small">Até</label>
							<input type="time" class="form-control" name="availability[wednesday][end]" required>
						</div>
					</div>
				</div>
				<div class="mb-3">
					<label for="thursday" class="form-label">Quinta-Feira</label>
					<div class="row g-2">
						<div class="col">
							<label class="form-label small">Das</label>
							<input type="time" class="form-control" name="availability[thursday][start]" required>
						</div>
						<div class="col">
							<label class="form-label small">Até</label>
							<input type="time" class="form-control" name="availability[thursday][end]" required>
						</div>
					</div>
				</div>
				<div class="mb-3">
					<label for="friday" class="form-label">Sexta-Feira</label>
					<div class="row g-2">
						<div class="col">
							<label class="form-label small">Das</label>
							<input type="time" class="form-control" name="availability[friday][start]" required>
						</div>
						<div class="col">
							<label class="form-label small">Até</label>
							<input type="time" class="form-control" name="availability[friday][end]" required>
						</div>
					</div>
				</div>
				<div class="mb-3">
					<label for="saturday" class="form-label">Sábado</label>
					<div class="row g-2">
						<div class="col">
							<label class="form-label small">Das</label>
							<input type="time" class="form-control" name="availability[saturday][start]" required>
						</div>
						<div class="col">
							<label class="form-label small">Até</label>
							<input type="time" class="form-control" name="availability[saturday][end]" required>
						</div>
					</div>
				</div>
				<div class="mb-3">
					<label for="sunday" class="form-label">Domingo</label>
					<div class="row g-2">
						<div class="col">
							<label class="form-label small">Das</label>
							<input type="time" class="form-control" name="availability[sunday][start]" required>
						</div>
						<div class="col">
							<label class="form-label small">Até</label>
							<input type="time" class="form-control" name="availability[sunday][end]" required>
						</div>
					</div>
				</div>
			</div>`
		} else {
			// Campo vazio para Supervisando
			html = ``
		}

		// Retorne o HTML apropriado
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}
}

func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		err := json.NewDecoder(r.Body).Decode(&credentials)
		if err != nil {
			sendResponse(w, http.StatusBadRequest, "Dados inválidos.")
			return
		}

		var id int
		var hashedPassword string
		err = db.QueryRow("SELECT id, password_hash FROM users WHERE email = $1", credentials.Email).Scan(&id, &hashedPassword)
		if err != nil {
			sendResponse(w, http.StatusUnauthorized, "Usuário ou senha inválidos.")
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password))
		if err != nil {
			sendResponse(w, http.StatusUnauthorized, "Usuário ou senha inválidos.")
			return
		}

		session, err := sessions.GetSession(r)
		if err != nil {
			sendResponse(w, http.StatusInternalServerError, "Erro ao recuperar a sessão.")
			return
		}

		session.Values["user_id"] = id
		session.Save(r, w)

		sendResponse(w, http.StatusOK, "Login realizado com sucesso!")
	}
}
