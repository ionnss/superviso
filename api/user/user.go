package user

import (
	"database/sql"
	"net/http"
	"superviso/api/auth"
	"superviso/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User

		// Parse do form
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`<div class="alert alert-danger">Erro ao processar formulário</div>`))
			return
		}

		// Preenche struct do usuário
		user.FirstName = r.FormValue("first_name")
		user.LastName = r.FormValue("last_name")
		user.Email = r.FormValue("email")
		user.CPF = r.FormValue("cpf")
		user.CRP = r.FormValue("crp")
		user.TheoryApproach = r.FormValue("theory_approach")

		// Hash da senha
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`<div class="alert alert-danger">Erro ao processar senha</div>`))
			return
		}
		user.PasswordHash = string(passwordHash)

		// Insere usuário no banco
		query := `
			INSERT INTO users (first_name, last_name, cpf, email, password_hash, crp, theory_approach, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING id`

		err = db.QueryRow(
			query,
			user.FirstName,
			user.LastName,
			user.CPF,
			user.Email,
			user.PasswordHash,
			user.CRP,
			user.TheoryApproach,
			time.Now(),
		).Scan(&user.ID)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`<div class="alert alert-danger">Erro ao cadastrar usuário</div>`))
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`<div class="alert alert-success">Usuário cadastrado com sucesso! Redirecionando...</div>`))
	}
}

func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User

		email := r.FormValue("email")
		password := r.FormValue("password")

		// Busca usuário pelo email
		query := `SELECT id, email, password_hash FROM users WHERE email = $1`
		err := db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.PasswordHash)

		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`<div class="alert alert-danger">Usuário não encontrado</div>`))
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`<div class="alert alert-danger">Erro ao buscar usuário</div>`))
			return
		}

		// Verifica senha
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`<div class="alert alert-danger">Senha incorreta</div>`))
			return
		}

		// Gera o token
		token, err := auth.GenerateToken(user.ID, user.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`<div class="alert alert-danger">Erro ao gerar token</div>`))
			return
		}

		// Define o cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true, // Em produção
			SameSite: http.SameSiteStrictMode,
			MaxAge:   24 * 60 * 60, // 24 horas
		})

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<div class="alert alert-success">Login realizado com sucesso! Redirecionando...</div>`))
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Remove o cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Se a requisição veio via HTMX, retorna um header especial
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/?msg=logout_success")
	} else {
		// Caso contrário, faz redirecionamento normal
		http.Redirect(w, r, "/?msg=logout_success", http.StatusSeeOther)
	}
}
