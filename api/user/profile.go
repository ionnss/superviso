package user

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"superviso/api/auth"
	"text/template"
)

func UpdateProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		// Atualiza informações básicas
		_, err := db.Exec(`
			UPDATE users 
			SET first_name = $1, last_name = $2
			WHERE id = $3`,
			r.FormValue("first_name"),
			r.FormValue("last_name"),
			userID,
		)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`<div class="alert alert-danger">Erro ao atualizar perfil</div>`))
			return
		}

		// Verifica se o modo supervisor está ativo
		if r.FormValue("is_supervisor") == "on" {
			// Validação do preço
			if r.FormValue("session_price") == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`<div class="alert alert-danger">O valor da sessão é obrigatório</div>`))
				return
			}

			// Validação dos dias
			if len(r.Form["available_days"]) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`<div class="alert alert-danger">Selecione pelo menos um dia disponível</div>`))
				return
			}

			// Validação dos horários
			if r.FormValue("start_time") == "" || r.FormValue("end_time") == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`<div class="alert alert-danger">Os horários são obrigatórios</div>`))
				return
			}

			availableDays := strings.Join(r.Form["available_days"], ",")

			_, err = db.Exec(`
				INSERT INTO supervisor_profiles 
				(user_id, session_price, available_days, start_time, end_time)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (user_id) 
				DO UPDATE SET 
					session_price = $2,
					available_days = $3,
					start_time = $4,
					end_time = $5`,
				userID,
				r.FormValue("session_price"),
				availableDays,
				r.FormValue("start_time"),
				r.FormValue("end_time"),
			)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`<div class="alert alert-danger">Erro ao atualizar perfil de supervisor</div>`))
				return
			}
		} else {
			// Se não está ativo, remove o perfil de supervisor se existir
			_, err = db.Exec(`DELETE FROM supervisor_profiles WHERE user_id = $1`, userID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`<div class="alert alert-danger">Erro ao atualizar perfil</div>`))
				return
			}
		}

		w.Write([]byte(`<div class="alert alert-success">Perfil atualizado com sucesso!</div>`))
	}
}

// Adicione esta função
func contains(list string, item string) bool {
	return strings.Contains(list, item)
}

// Modifique GetProfile para incluir a função no template
func GetProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		var user struct {
			FirstName     string
			LastName      string
			Email         string
			IsSupervisor  bool
			HasRole       bool
			SessionPrice  float64
			AvailableDays string
			StartTime     string
			EndTime       string
		}

		// Busca dados básicos
		err := db.QueryRow(`
			SELECT first_name, last_name, email 
			FROM users WHERE id = $1`,
			userID,
		).Scan(&user.FirstName, &user.LastName, &user.Email)

		if err != nil {
			http.Error(w, "Erro ao buscar dados do usuário", http.StatusInternalServerError)
			return
		}

		// Busca dados de supervisor se existirem
		err = db.QueryRow(`
			SELECT session_price, available_days, start_time, end_time 
			FROM supervisor_profiles 
			WHERE user_id = $1`,
			userID,
		).Scan(&user.SessionPrice, &user.AvailableDays, &user.StartTime, &user.EndTime)

		if err != sql.ErrNoRows {
			user.IsSupervisor = true
		}

		// Verifica se o usuário já tem um papel definido
		var roleExists bool
		err = db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM supervisor_profiles WHERE user_id = $1
			)`, userID).Scan(&roleExists)

		if err != nil {
			http.Error(w, "Erro ao verificar papel do usuário", http.StatusInternalServerError)
			return
		}

		user.HasRole = roleExists

		// Adiciona função helper ao template
		funcMap := template.FuncMap{
			"contains": contains,
		}

		tmpl := template.Must(template.New("profile.html").Funcs(funcMap).ParseFiles("view/profile.html"))
		tmpl.Execute(w, user)
	}
}

// Adicionar esta função
func ToggleSupervisor(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		// Verifica se já é supervisor
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM supervisor_profiles WHERE user_id = $1)", userID).Scan(&exists)
		if err != nil {
			http.Error(w, "Erro ao verificar perfil", http.StatusInternalServerError)
			return
		}

		if exists {
			// Se já existe, retorna os dados atuais
			var profile struct {
				SessionPrice  float64
				AvailableDays string
				StartTime     string
				EndTime       string
			}

			err := db.QueryRow(`
				SELECT session_price, available_days, start_time, end_time 
				FROM supervisor_profiles 
				WHERE user_id = $1`,
				userID,
			).Scan(&profile.SessionPrice, &profile.AvailableDays, &profile.StartTime, &profile.EndTime)

			if err != nil {
				http.Error(w, "Erro ao buscar dados", http.StatusInternalServerError)
				return
			}

			// Adiciona função helper ao template
			funcMap := template.FuncMap{
				"contains": contains,
			}

			tmpl := template.Must(template.New("supervisor_fields.html").Funcs(funcMap).ParseFiles("view/partials/supervisor_fields.html"))
			tmpl.Execute(w, profile)
		} else {
			// Se não existe, retorna o template vazio com as funções
			funcMap := template.FuncMap{
				"contains": contains,
			}

			tmpl := template.Must(template.New("supervisor_fields.html").Funcs(funcMap).ParseFiles("view/partials/supervisor_fields.html"))
			tmpl.Execute(w, nil)
		}
	}
}

// Adicionar esta função
func CheckUserRole(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		var hasRole bool
		err := db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM supervisor_profiles WHERE user_id = $1
			)`, userID).Scan(&hasRole)

		if err != nil {
			http.Error(w, "Erro ao verificar papel do usuário", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"hasRole": hasRole})
	}
}
