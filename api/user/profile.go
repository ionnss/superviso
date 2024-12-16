package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"superviso/api/auth"
	"superviso/models"
	"text/template"
	"time"
)

func UpdateProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)
		var updateSuccess bool = false

		// Atualiza informações básicas
		_, err := db.Exec(`
				UPDATE users 
				SET first_name = $1, 
					last_name = $2,
					crp = $3,
					theory_approach = $4
				WHERE id = $5`,
			r.FormValue("first_name"),
			r.FormValue("last_name"),
			r.FormValue("crp"),
			r.FormValue("theory_approach"),
			userID,
		)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`<div class="alert alert-danger">Erro ao atualizar perfil</div>`))
			return
		}

		updateSuccess = true // Marca que houve atualização básica com sucesso

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

			updateSuccess = true // Marca que houve atualização do supervisor com sucesso
		} else {
			// Se não está ativo, remove o perfil de supervisor se existir
			_, err = db.Exec(`DELETE FROM supervisor_profiles WHERE user_id = $1`, userID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`<div class="alert alert-danger">Erro ao atualizar perfil</div>`))
				return
			}
		}

		// Se qualquer atualização foi bem sucedida, mostra mensagem de sucesso
		if updateSuccess {
			w.Write([]byte(`<div class="alert alert-success">Perfil atualizado com sucesso!</div>`))
		}
	}
}

// Atualizar GetProfile para incluir os dados corretos
func GetProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		// Buscar dados do usuário
		var user models.User
		err := db.QueryRow(`
			SELECT first_name, last_name, email, crp, theory_approach 
			FROM users WHERE id = $1`,
			userID).Scan(&user.FirstName, &user.LastName, &user.Email, &user.CRP, &user.TheoryApproach)
		if err != nil {
			http.Error(w, "Erro ao buscar usuário", http.StatusInternalServerError)
			return
		}

		// Verificar se é supervisor
		var isSupervisor bool
		err = db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM supervisor_profiles WHERE user_id = $1
			)`, userID).Scan(&isSupervisor)
		if err != nil {
			http.Error(w, "Erro ao verificar perfil", http.StatusInternalServerError)
			return
		}

		// Preparar dados para o template
		data := struct {
			User         models.User
			IsSupervisor bool
			SessionPrice float64
			WeeklyHours  []models.SupervisorWeeklyHours
			WeekDays     []int
		}{
			User:         user,
			IsSupervisor: isSupervisor,
			WeekDays:     []int{1, 2, 3, 4, 5, 6, 7},
		}

		if isSupervisor {
			// Buscar preço da sessão
			err = db.QueryRow(`
				SELECT session_price FROM supervisor_profiles 
				WHERE user_id = $1`, userID).Scan(&data.SessionPrice)
			if err != nil && err != sql.ErrNoRows {
				http.Error(w, "Erro ao buscar dados de supervisor", http.StatusInternalServerError)
				return
			}

			// Buscar horários semanais
			rows, err := db.Query(`
				SELECT weekday, start_time, end_time 
				FROM supervisor_weekly_hours 
				WHERE supervisor_id = $1 
				ORDER BY weekday`, userID)
			if err != nil {
				http.Error(w, "Erro ao buscar horários", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var h models.SupervisorWeeklyHours
				err := rows.Scan(&h.Weekday, &h.StartTime, &h.EndTime)
				if err != nil {
					http.Error(w, "Erro ao ler horários", http.StatusInternalServerError)
					return
				}
				data.WeeklyHours = append(data.WeeklyHours, h)
			}
		}

		// Adicionar funções ao template
		funcMap := template.FuncMap{
			"formatWeekday": func(day int) string {
				weekdays := map[int]string{
					1: "Segunda",
					2: "Terça",
					3: "Quarta",
					4: "Quinta",
					5: "Sexta",
					6: "Sábado",
					7: "Domingo",
				}
				return weekdays[day]
			},
		}

		tmpl := template.Must(template.New("profile.html").
			Funcs(funcMap).
			ParseFiles("view/profile.html"))
		tmpl.Execute(w, data)
	}
}

// Adicionar esta função
func ToggleSupervisor(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)
		isActivating := r.FormValue("is_supervisor") == "on"

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Erro ao iniciar transação", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		if isActivating {
			// Criar perfil de supervisor
			_, err = tx.Exec(`
				INSERT INTO supervisor_profiles (user_id, user_crp, session_price)
				VALUES ($1, (SELECT crp FROM users WHERE id = $1), 0)
				ON CONFLICT (user_id) DO NOTHING`,
				userID)
		} else {
			// Remover perfil de supervisor
			_, err = tx.Exec(`
				DELETE FROM supervisor_profiles 
				WHERE user_id = $1`,
				userID)
		}

		if err != nil {
			http.Error(w, "Erro ao atualizar perfil", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Erro ao finalizar alterações", http.StatusInternalServerError)
			return
		}

		// Retornar os campos de supervisor se ativado
		if isActivating {
			// Buscar dados atualizados do supervisor
			type WeeklyHour struct {
				StartTime string
				EndTime   string
			}

			var data struct {
				SessionPrice        float64
				WeeklyHours         map[int]WeeklyHour
				WeekDays            []int
				AvailabilityStart   string
				AvailabilityEnd     string
				AvailabilityPeriods []models.SupervisorAvailabilityPeriod
			}

			data.WeeklyHours = make(map[int]WeeklyHour)

			// Buscar horários semanais se existirem
			rows, err := db.Query(`
				SELECT weekday, start_time, end_time 
				FROM supervisor_weekly_hours 
				WHERE supervisor_id = $1 
				ORDER BY weekday`, userID)
			if err != nil {
				http.Error(w, "Erro ao buscar horários", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var weekday int
				var h WeeklyHour
				err := rows.Scan(&weekday, &h.StartTime, &h.EndTime)
				if err != nil {
					http.Error(w, "Erro ao ler horários", http.StatusInternalServerError)
					return
				}
				data.WeeklyHours[weekday] = h
			}

			// Adicionar os dias da semana
			data.WeekDays = []int{1, 2, 3, 4, 5, 6, 7}

			// Formatar data atual para o valor mínimo dos inputs de data
			now := time.Now().Format("2006-01-02")
			data.AvailabilityStart = now
			data.AvailabilityEnd = now

			// Buscar períodos de disponibilidade
			rows, err = db.Query(`
				SELECT id, start_date, end_date 
				FROM supervisor_availability_periods 
				WHERE supervisor_id = $1 
				AND end_date >= CURRENT_DATE
				ORDER BY start_date`, userID)
			if err != nil {
				http.Error(w, "Erro ao buscar períodos", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var p models.SupervisorAvailabilityPeriod
				err := rows.Scan(&p.ID, &p.StartDate, &p.EndDate)
				if err != nil {
					http.Error(w, "Erro ao ler períodos", http.StatusInternalServerError)
					return
				}
				data.AvailabilityPeriods = append(data.AvailabilityPeriods, p)
			}

			// Renderizar template com os dados
			funcMap := template.FuncMap{
				"formatWeekday": func(day int) string {
					weekdays := map[int]string{
						1: "Segunda",
						2: "Terça",
						3: "Quarta",
						4: "Quinta",
						5: "Sexta",
						6: "Sábado",
						7: "Domingo",
					}
					return weekdays[day]
				},
				"now": func() string {
					return time.Now().Format("2006-01-02")
				},
			}

			tmpl := template.Must(template.New("supervisor_fields.html").
				Funcs(funcMap).
				ParseFiles("view/partials/supervisor_fields.html"))
			tmpl.Execute(w, data)
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

func UploadProfileImage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		// Parse do formulário multipart
		err := r.ParseMultipartForm(10 << 20) // 10MB max
		if err != nil {
			w.Write([]byte(`<div class="alert alert-danger">Erro ao processar imagem</div>`))
			return
		}

		file, handler, err := r.FormFile("profile_image")
		if err != nil {
			w.Write([]byte(`<div class="alert alert-danger">Erro ao receber arquivo</div>`))
			return
		}
		defer file.Close()

		// Validar tipo do arquivo
		if !strings.HasPrefix(handler.Header.Get("Content-Type"), "image/") {
			w.Write([]byte(`<div class="alert alert-danger">Arquivo deve ser uma imagem</div>`))
			return
		}

		// Criar diretório se não existir
		uploadDir := "uploads/profile_images"
		os.MkdirAll(uploadDir, 0755)

		// Gerar nome único para o arquivo
		filename := fmt.Sprintf("%d_%s", userID, handler.Filename)
		filepath := fmt.Sprintf("%s/%s", uploadDir, filename)

		// Salvar arquivo
		dst, err := os.Create(filepath)
		if err != nil {
			w.Write([]byte(`<div class="alert alert-danger">Erro ao salvar imagem</div>`))
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			w.Write([]byte(`<div class="alert alert-danger">Erro ao copiar arquivo</div>`))
			return
		}

		// Atualizar caminho no banco
		_, err = db.Exec(`
			UPDATE users 
			SET profile_image = $1 
			WHERE id = $2`,
			"/"+filepath, userID)

		if err != nil {
			w.Write([]byte(`<div class="alert alert-danger">Erro ao atualizar perfil</div>`))
			return
		}

		w.Write([]byte(`<div class="alert alert-success">Foto atualizada com sucesso!</div>`))
	}
}
