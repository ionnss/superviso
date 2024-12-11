package appointment

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"superviso/api/auth"
	"superviso/models"
	"time"
)

// GetNewAppointmentForm renderiza o formulário de novo agendamento
func GetNewAppointmentForm(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obter ID do supervisor da URL
		supervisorID := r.URL.Query().Get("supervisor_id")
		if supervisorID == "" {
			http.Error(w, "Supervisor não especificado", http.StatusBadRequest)
			return
		}

		// Buscar dados do supervisor
		var supervisor models.Supervisor
		err := db.QueryRow(`
			SELECT u.id, u.first_name, u.last_name, u.crp, sp.available_days
			FROM users u
			JOIN supervisor_profiles sp ON u.id = sp.user_id
			WHERE u.id = $1`,
			supervisorID).Scan(
			&supervisor.UserID,
			&supervisor.FirstName,
			&supervisor.LastName,
			&supervisor.CRP,
			&supervisor.AvailableDays,
		)
		if err != nil {
			http.Error(w, "Erro ao buscar supervisor", http.StatusInternalServerError)
			return
		}

		// Preparar dados para o template
		data := struct {
			Supervisor    models.Supervisor
			AvailableDays []string
		}{
			Supervisor:    supervisor,
			AvailableDays: parseAvailableDays(supervisor.AvailableDays),
		}

		// Funções para o template
		funcMap := template.FuncMap{
			"formatWeekday": formatWeekday,
			"formatWeekdayFromDate": func(t time.Time) string {
				weekdays := map[int]string{
					0: "Domingo",
					1: "Segunda-feira",
					2: "Terça-feira",
					3: "Quarta-feira",
					4: "Quinta-feira",
					5: "Sexta-feira",
					6: "Sábado",
				}
				return weekdays[int(t.Weekday())]
			},
		}

		// Renderizar template
		tmpl := template.Must(template.New("schedule.html").
			Funcs(funcMap).
			ParseFiles("view/appointments/schedule.html"))

		tmpl.Execute(w, data)
	}
}

// GetAvailableSlots retorna os slots disponíveis para um dia específico
func GetAvailableSlots(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		supervisorID := r.URL.Query().Get("supervisor_id")
		day := r.URL.Query().Get("day")

		// Converter supervisorID para int
		supID, err := strconv.Atoi(supervisorID)
		if err != nil {
			http.Error(w, "ID do supervisor inválido", http.StatusBadRequest)
			return
		}

		// Buscar horários do supervisor
		var startTime, endTime string
		err = db.QueryRow(`
			SELECT 
				start_time::time::text,
				end_time::time::text
			FROM supervisor_profiles
			WHERE user_id = $1`,
			supID).Scan(&startTime, &endTime)
		if err != nil {
			http.Error(w, "Erro ao buscar horários do supervisor", http.StatusInternalServerError)
			return
		}

		// Gerar slots para o dia específico
		allSlots := generateSlots(supID, startTime, endTime, []string{day})

		// Primeiro, inserir os slots no banco se não existirem
		for _, slot := range allSlots {
			_, err = db.Exec(`
				INSERT INTO available_slots 
				(supervisor_id, slot_date, start_time, end_time, status)
				VALUES ($1, $2, $3::time, $4::time, $5)
				ON CONFLICT (supervisor_id, slot_date, start_time) DO NOTHING`,
				slot.SupervisorID, slot.SlotDate, slot.StartTime, slot.EndTime, slot.Status)
			if err != nil {
				log.Printf("Erro ao inserir slot: %v", err)
				continue
			}
		}

		// Filtrar slots já agendados
		rows, err := db.Query(`
			SELECT slot_date::date
			FROM available_slots
			WHERE supervisor_id = $1
			AND status = 'booked'`,
			supID)
		if err != nil {
			http.Error(w, "Erro ao verificar slots agendados", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Criar mapa de datas ocupadas
		bookedDates := make(map[string]bool)
		for rows.Next() {
			var date time.Time
			err := rows.Scan(&date)
			if err != nil {
				continue
			}
			bookedDates[date.Format("2006-01-02")] = true
		}

		// Filtrar slots disponíveis
		var availableSlots []models.AvailableSlot
		for _, slot := range allSlots {
			dateStr := slot.SlotDate.Format("2006-01-02")
			if !bookedDates[dateStr] {
				availableSlots = append(availableSlots, slot)
			}
		}

		// Funções para o template
		funcMap := template.FuncMap{
			"formatDate":           formatDate,
			"formatTimeForDisplay": formatTimeForDisplay,
			"formatWeekdayFromDate": func(t time.Time) string {
				weekdays := map[int]string{
					0: "Domingo",
					1: "Segunda-feira",
					2: "Terça-feira",
					3: "Quarta-feira",
					4: "Quinta-feira",
					5: "Sexta-feira",
					6: "Sábado",
				}
				return weekdays[int(t.Weekday())]
			},
		}

		// Renderizar partial com os slots
		tmpl := template.Must(template.New("slots_grid.html").
			Funcs(funcMap).
			ParseFiles("view/appointments/partials/slots_grid.html"))
		tmpl.Execute(w, availableSlots)
	}
}

// Funções auxiliares
func formatDate(t time.Time) string {
	return t.Format("02/01/2006")
}

func parseAvailableDays(days string) []string {
	if days == "" {
		return []string{}
	}
	return strings.Split(days, ",")
}

// BookAppointment processa o agendamento de uma supervisão
func BookAppointment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Iniciando processo de agendamento...")

		// Obter ID do usuário do contexto
		userID := r.Context().Value(auth.UserIDKey).(int)
		log.Printf("UserID do contexto: %d", userID)

		// Obter ID do slot
		slotID, err := strconv.Atoi(r.FormValue("slot_id"))
		log.Printf("SlotID recebido: %d", slotID)
		if err != nil {
			log.Printf("Erro ao converter slot_id: %v", err)
			http.Error(w, "ID do slot inválido", http.StatusBadRequest)
			return
		}

		// Verificar se o usuário já tem um agendamento pendente com este supervisor
		var existingAppointment bool
		err = db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM appointments a
				JOIN available_slots s ON a.slot_id = s.id
				WHERE a.supervisee_id = $1 
				AND s.supervisor_id = (SELECT supervisor_id FROM available_slots WHERE id = $2)
				AND a.status = 'pending'
			)`, userID, slotID).Scan(&existingAppointment)

		if err != nil {
			http.Error(w, "Erro ao verificar agendamentos existentes", http.StatusInternalServerError)
			return
		}

		if existingAppointment {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`
				<div class="alert alert-warning">
					<i class="fas fa-exclamation-circle me-2"></i>
					Você já possui um agendamento pendente com este supervisor.
					Aguarde a confirmação ou cancele o agendamento anterior.
				</div>
			`))
			return
		}

		// Iniciar transação
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Erro ao iniciar transação", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Verificar se o slot está disponível
		var supervisorID int
		err = tx.QueryRow(`
			SELECT supervisor_id 
			FROM available_slots 
			WHERE id = $1 AND status = 'available'
			FOR UPDATE`,
			slotID).Scan(&supervisorID)

		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`
				<div class="alert alert-warning">
					<i class="fas fa-exclamation-circle me-2"></i>
					Este horário não está mais disponível.
				</div>
			`))
			return
		}
		if err != nil {
			http.Error(w, "Erro ao verificar disponibilidade", http.StatusInternalServerError)
			return
		}

		// Criar agendamento
		_, err = tx.Exec(`
			INSERT INTO appointments 
			(supervisor_id, supervisee_id, slot_id, status) 
			VALUES ($1, $2, $3, 'pending')`,
			supervisorID, userID, slotID)
		if err != nil {
			http.Error(w, "Erro ao criar agendamento", http.StatusInternalServerError)
			return
		}

		// Atualizar status do slot
		_, err = tx.Exec(`
			UPDATE available_slots 
			SET status = 'booked' 
			WHERE id = $1`,
			slotID)
		if err != nil {
			http.Error(w, "Erro ao atualizar slot", http.StatusInternalServerError)
			return
		}

		// Commit da transação
		if err := tx.Commit(); err != nil {
			http.Error(w, "Erro ao finalizar agendamento", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(`
			<div class="alert alert-success">
				<i class="fas fa-check-circle me-2"></i>
				Agendamento realizado com sucesso! Aguardando confirmação do supervisor.
			</div>
		`))
	}
}

func formatWeekday(day string) string {
	weekdays := map[string]string{
		"1": "Segunda",
		"2": "Terça",
		"3": "Quarta",
		"4": "Quinta",
		"5": "Sexta",
		"6": "Sábado",
		"7": "Domingo",
	}
	return weekdays[day]
}

// GenerateSlots gera slots de horário para as próximas 4 semanas
func generateSlots(supervisorID int, startTime, endTime string, availableDays []string) []models.AvailableSlot {
	var slots []models.AvailableSlot
	now := time.Now()

	// Gerar slots para as próximas 4 semanas
	for week := 0; week < 4; week++ {
		for _, dayStr := range availableDays {
			day, _ := strconv.Atoi(dayStr)

			// Encontrar próxima data para este dia da semana
			date := nextWeekday(now.AddDate(0, 0, week*7), time.Weekday(day))

			// Criar slot
			slot := models.AvailableSlot{
				SupervisorID: supervisorID,
				SlotDate:     date,
				StartTime:    formatTimeForDB(startTime),
				EndTime:      formatTimeForDB(endTime),
				Status:       "available",
			}
			slots = append(slots, slot)
		}
	}
	return slots
}

// nextWeekday retorna a próxima data para um dia específico da semana
func nextWeekday(start time.Time, weekday time.Weekday) time.Time {
	date := start
	for date.Weekday() != weekday {
		date = date.AddDate(0, 0, 1)
	}
	return date
}

// formatTimeForDB converte o horário para o formato aceito pelo PostgreSQL
func formatTimeForDB(timeStr string) string {
	return timeStr // O horário já vem no formato correto do banco
}

// formatTimeForDisplay formata o horário para exibição
func formatTimeForDisplay(timeStr string) string {
	// Parse do horário no formato HH:MM:SS
	t, err := time.Parse("15:04:05", timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("15:04")
}
