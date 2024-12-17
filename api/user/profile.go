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
	"text/template"
	"time"
)

var funcMap = template.FuncMap{
	"contains": contains,
	"formatWeekday": func(day int) string {
		weekdays := map[int]string{
			0: "Domingo",
			1: "Segunda",
			2: "Terça",
			3: "Quarta",
			4: "Quinta",
			5: "Sexta",
			6: "Sábado",
		}
		return weekdays[day]
	},
	"now": func() string {
		return time.Now().Format("2006-01-02")
	},
}

func UpdateProfile(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Erro ao iniciar transação", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// 1. Atualizar dados básicos do usuário
		_, err = tx.Exec(`
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
			http.Error(w, "Erro ao atualizar perfil", http.StatusInternalServerError)
			return
		}

		// 2. Se for supervisor, atualizar configurações
		if r.FormValue("is_supervisor") == "on" {
			// Atualizar perfil de supervisor
			_, err = tx.Exec(`
				INSERT INTO supervisor_profiles (user_id, session_price)
				VALUES ($1, $2)
				ON CONFLICT (user_id) DO UPDATE 
				SET session_price = $2`,
				userID, r.FormValue("session_price"))
			if err != nil {
				http.Error(w, "Erro ao atualizar valor da sessão", http.StatusInternalServerError)
				return
			}

			// Atualizar período de disponibilidade
			_, err = tx.Exec(`
				INSERT INTO supervisor_availability_periods 
				(supervisor_id, start_date, end_date)
				VALUES ($1, $2::date, $3::date)
				ON CONFLICT ON CONSTRAINT unique_supervisor_period 
				DO UPDATE SET 
					start_date = EXCLUDED.start_date,
					end_date = EXCLUDED.end_date`,
				userID,
				r.FormValue("availability_start"),
				r.FormValue("availability_end"))
			if err != nil {
				http.Error(w, "Erro ao atualizar período", http.StatusInternalServerError)
				return
			}

			// Limpar horários antigos
			_, err = tx.Exec(`DELETE FROM supervisor_weekly_hours WHERE supervisor_id = $1`, userID)
			if err != nil {
				http.Error(w, "Erro ao limpar horários", http.StatusInternalServerError)
				return
			}

			// Inserir novos horários
			for day := 0; day <= 6; day++ {
				if r.FormValue(fmt.Sprintf("day_%d", day)) == "on" {
					startTime := r.FormValue(fmt.Sprintf("start_time_%d", day))
					endTime := r.FormValue(fmt.Sprintf("end_time_%d", day))

					if startTime != "" && endTime != "" {
						_, err = tx.Exec(`
							INSERT INTO supervisor_weekly_hours 
							(supervisor_id, weekday, start_time, end_time)
							VALUES ($1, $2, $3, $4)`,
							userID, day, startTime, endTime)

						if err != nil {
							http.Error(w, "Erro ao salvar horários", http.StatusInternalServerError)
							return
						}
					}
				}
			}

			// Gerar slots para todo o período de disponibilidade
			startDate, _ := time.Parse("2006-01-02", r.FormValue("availability_start"))
			endDate, _ := time.Parse("2006-01-02", r.FormValue("availability_end"))

			// Para cada dia selecionado
			for day := 0; day <= 6; day++ {
				if r.FormValue(fmt.Sprintf("day_%d", day)) == "on" {
					startTime := r.FormValue(fmt.Sprintf("start_time_%d", day))
					endTime := r.FormValue(fmt.Sprintf("end_time_%d", day))

					// Começar do primeiro dia do período
					currentDate := startDate

					// Enquanto não passar do fim do período
					for currentDate.Before(endDate) || currentDate.Equal(endDate) {
						// Se for o dia da semana correto
						weekday := int(currentDate.Weekday())
						if weekday == day {
							_, err = tx.Exec(`
								INSERT INTO available_slots 
								(supervisor_id, slot_date, start_time, end_time, status)
								VALUES ($1, $2::date, $3::time, $4::time, 'available')
								ON CONFLICT (supervisor_id, slot_date, start_time) DO NOTHING`,
								userID, currentDate, startTime, endTime)
							if err != nil {
								http.Error(w, "Erro ao gerar slots", http.StatusInternalServerError)
								return
							}
						}
						// Avançar para o próximo dia
						currentDate = currentDate.AddDate(0, 0, 1)
					}
				}
			}
		} else {
			// Se não está ativo, remove todos os dados do supervisor em uma transação
			_, err = tx.Exec(`
				WITH deleted_supervisor AS (
					DELETE FROM supervisor_profiles 
					WHERE user_id = $1 
					RETURNING user_id
				)
				DELETE FROM supervisor_weekly_hours 
				WHERE supervisor_id IN (SELECT user_id FROM deleted_supervisor)`,
				userID)
			if err != nil {
				http.Error(w, "Erro ao remover perfil de supervisor", http.StatusInternalServerError)
				return
			}

			// Remover períodos de disponibilidade
			_, err = tx.Exec(`
				DELETE FROM supervisor_availability_periods 
				WHERE supervisor_id = $1`,
				userID)
			if err != nil {
				http.Error(w, "Erro ao remover períodos", http.StatusInternalServerError)
				return
			}

			// Remover slots disponíveis
			_, err = tx.Exec(`
				DELETE FROM available_slots 
				WHERE supervisor_id = $1`,
				userID)
			if err != nil {
				http.Error(w, "Erro ao remover slots", http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "Erro ao finalizar alterações", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(`
			<div class="alert alert-success">
				<i class="fas fa-check-circle me-2"></i>
				Perfil atualizado com sucesso!
			</div>
		`))
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
			FirstName      string
			LastName       string
			Email          string
			CRP            string
			TheoryApproach string
			IsSupervisor   bool
			HasRole        bool
			SessionPrice   float64
			AvailableDays  string
			StartTime      string
			EndTime        string
		}

		// Busca dados básicos
		err := db.QueryRow(`
			SELECT first_name, last_name, email, crp, theory_approach 
			FROM users WHERE id = $1`,
			userID,
		).Scan(&user.FirstName, &user.LastName, &user.Email, &user.CRP, &user.TheoryApproach)

		if err != nil {
			http.Error(w, "Erro ao buscar dados do usuário", http.StatusInternalServerError)
			return
		}

		// Busca dados de supervisor se existirem
		err = db.QueryRow(`
			SELECT session_price
			FROM supervisor_profiles 
			WHERE user_id = $1`,
			userID,
		).Scan(&user.SessionPrice)

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
				SessionPrice float64
				StartDate    string
				EndDate      string
				WeeklyHours  map[int]struct {
					StartTime string
					EndTime   string
				}
			}
			profile.WeeklyHours = make(map[int]struct{ StartTime, EndTime string })

			// Buscar valor da sessão
			err := db.QueryRow(`
				SELECT session_price 
				FROM supervisor_profiles 
				WHERE user_id = $1`,
				userID).Scan(&profile.SessionPrice)

			if err != nil {
				http.Error(w, "Erro ao buscar dados", http.StatusInternalServerError)
				return
			}

			// Buscar período
			err = db.QueryRow(`
				SELECT TO_CHAR(start_date, 'YYYY-MM-DD'), TO_CHAR(end_date, 'YYYY-MM-DD')
				FROM supervisor_availability_periods 
				WHERE supervisor_id = $1`,
				userID).Scan(&profile.StartDate, &profile.EndDate)

			if err != nil && err != sql.ErrNoRows {
				http.Error(w, "Erro ao buscar período", http.StatusInternalServerError)
				return
			}

			// Buscar horários
			rows, err := db.Query(`
				SELECT weekday, start_time, end_time 
				FROM supervisor_weekly_hours 
				WHERE supervisor_id = $1`,
				userID)
			if err != nil {
				http.Error(w, "Erro ao buscar horários", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			for rows.Next() {
				var day int
				var start, end string
				if err := rows.Scan(&day, &start, &end); err != nil {
					http.Error(w, "Erro ao ler horários", http.StatusInternalServerError)
					return
				}
				profile.WeeklyHours[day] = struct{ StartTime, EndTime string }{start, end}
			}

			// Renderizar template
			tmpl := template.Must(template.New("supervisor_fields.html").
				Funcs(funcMap).
				ParseFiles("view/partials/supervisor_fields.html"))
			tmpl.Execute(w, profile)
		} else {
			// Se não existe, retorna o template vazio com as funções
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
