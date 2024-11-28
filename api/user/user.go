// superviso/api/user/user.go
package user

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"superviso/models"
)

func Register(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			http.Error(w, "Erro ao decodificar dados", http.StatusBadRequest)
			return
		}

		query := `
			INSERT INTO users (first_name, last_name, email, password_hash, crp, cpf, theory_approach, user_role)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err = db.Exec(query, user.FirstName, user.LastName, user.Email, user.PasswordHash, user.CRP, user.CPF, user.TheoryApproach, user.UserRole)
		if err != nil {
			http.Error(w, "Erro ao salvar usuário", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Usuário registrado com sucesso!"))
	}
}
