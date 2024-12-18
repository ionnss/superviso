package appointment

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"superviso/api/auth"
	"superviso/models"
	"superviso/utils"
	"time"
)

var funcMap = template.FuncMap{
	"formatWeekday": utils.FormatWeekday,
	"formatDate":    utils.FormatDate,
	"formatTime":    utils.FormatTime,
	"now": func() string {
		return time.Now().Format("2006-01-02")
	},
	"formatDateISO": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
	"groupSlotsByDate": func(slots []models.AvailableSlot) map[time.Time][]models.AvailableSlot {
		grouped := make(map[time.Time][]models.AvailableSlot)
		for _, slot := range slots {
			date := slot.SlotDate
			grouped[date] = append(grouped[date], slot)
		}
		return grouped
	},
}

// GetNewAppointmentForm renderiza o formulário de novo agendamento
func GetNewAppointmentForm(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		supervisorID := r.URL.Query().Get("supervisor_id")
		if supervisorID == "" {
			http.Error(w, "Supervisor não especificado", http.StatusBadRequest)
			return
		}

		// Buscar dados do supervisor e slots disponíveis
		var supervisor struct {
			ID             int
			FirstName      string
			LastName       string
			CRP            string
			TheoryApproach string
			SessionPrice   float64
			StartDate      time.Time
			EndDate        time.Time
			WeeklyHours    map[int]struct {
				StartTime string
				EndTime   string
			}
			AvailableSlots []models.AvailableSlot
		}
		supervisor.WeeklyHours = make(map[int]struct {
			StartTime string
			EndTime   string
		})

		// Buscar dados básicos
		err := db.QueryRow(`
			SELECT u.id, u.first_name, u.last_name, u.crp, u.theory_approach,
				   sp.session_price, sap.start_date, sap.end_date
			FROM users u
			JOIN supervisor_profiles sp ON u.id = sp.user_id
			JOIN supervisor_availability_periods sap ON u.id = sap.supervisor_id
			WHERE u.id = $1 AND sap.end_date >= CURRENT_DATE`,
			supervisorID).Scan(
			&supervisor.ID, &supervisor.FirstName, &supervisor.LastName,
			&supervisor.CRP, &supervisor.TheoryApproach, &supervisor.SessionPrice,
			&supervisor.StartDate, &supervisor.EndDate)

		if err != nil {
			http.Error(w, "Supervisor não encontrado", http.StatusNotFound)
			return
		}

		// Buscar horários semanais
		rows, err := db.Query(`
			SELECT 
				weekday,
				TO_CHAR(start_time, 'HH24:MI') as start_time,
				TO_CHAR(end_time, 'HH24:MI') as end_time
			FROM supervisor_weekly_hours 
			WHERE supervisor_id = $1
			ORDER BY weekday`,
			supervisorID)
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
			supervisor.WeeklyHours[day] = struct {
				StartTime string
				EndTime   string
			}{StartTime: start, EndTime: end}
		}

		// Buscar slots disponíveis
		slots, err := db.Query(`
			SELECT id, 
				   slot_date,
				   TO_CHAR(start_time, 'HH24:MI') as start_time,
				   TO_CHAR(end_time, 'HH24:MI') as end_time,
				   status
			FROM available_slots 
			WHERE supervisor_id = $1 
			AND slot_date >= CURRENT_DATE
			AND status = 'available'
			ORDER BY slot_date, start_time`,
			supervisorID)
		if err != nil {
			http.Error(w, "Erro ao buscar slots", http.StatusInternalServerError)
			return
		}
		defer slots.Close()

		for slots.Next() {
			var slot models.AvailableSlot
			if err := slots.Scan(&slot.SlotID, &slot.SlotDate, &slot.StartTime, &slot.EndTime, &slot.Status); err != nil {
				http.Error(w, "Erro ao ler slots", http.StatusInternalServerError)
				return
			}
			supervisor.AvailableSlots = append(supervisor.AvailableSlots, slot)
		}

		// Renderizar template
		tmpl := template.Must(template.New("schedule.html").
			Funcs(funcMap).
			ParseFiles("view/appointments/schedule.html"))

		w.Header().Set("Content-Type", "text/html")
		tmpl.Execute(w, map[string]interface{}{
			"Supervisor": supervisor,
		})
	}
}

// GetAvailableSlots retorna os slots disponíveis para um dia específico
func GetAvailableSlots(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		supervisorID := r.URL.Query().Get("supervisor_id")
		dateStr := r.URL.Query().Get("date")

		if dateStr == "" {
			http.Error(w, "Data não especificada", http.StatusBadRequest)
			return
		}

		// Converter data
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Data inválida", http.StatusBadRequest)
			return
		}

		// 1. Verificar se está dentro do período de disponibilidade
		var periodExists bool
		err = db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM supervisor_availability_periods
				WHERE supervisor_id = $1 
				AND $2::date BETWEEN start_date AND end_date
			)`, supervisorID, date).Scan(&periodExists)
		if err != nil {
			http.Error(w, "Erro ao verificar disponibilidade", http.StatusInternalServerError)
			return
		}
		if !periodExists {
			renderNoSlots(w, date)
			return
		}

		// 2. Verificar se é um dia que o supervisor atende
		weekday := int(date.Weekday())
		if weekday == 0 {
			weekday = 7
		} // Domingo = 7

		var startTime, endTime string
		err = db.QueryRow(`
			SELECT start_time, end_time 
			FROM supervisor_weekly_hours 
			WHERE supervisor_id = $1 AND weekday = $2`,
			supervisorID, weekday).Scan(&startTime, &endTime)
		if err == sql.ErrNoRows {
			renderNoSlots(w, date)
			return
		}
		if err != nil {
			http.Error(w, "Erro ao buscar horários", http.StatusInternalServerError)
			return
		}

		// 3. Gerar slots se não existirem
		_, err = db.Exec(`
			INSERT INTO available_slots 
				(supervisor_id, slot_date, start_time, end_time, status)
			SELECT 
				$1, $2::date, time, time + interval '1 hour', 'available'
			FROM generate_series(
				$3::time,
				$4::time - interval '1 hour',
					interval '1 hour'
			) as time
			ON CONFLICT (supervisor_id, slot_date, start_time) 
			DO NOTHING`,
			supervisorID, date, startTime, endTime)
		if err != nil {
			http.Error(w, "Erro ao gerar slots", http.StatusInternalServerError)
			return
		}

		// 4. Buscar slots disponíveis
		rows, err := db.Query(`
			SELECT id, start_time, end_time 
			FROM available_slots 
			WHERE supervisor_id = $1 
			AND slot_date = $2
			AND status = 'available'
			ORDER BY start_time`,
			supervisorID, date)
		if err != nil {
			http.Error(w, "Erro ao buscar slots", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var slots []models.AvailableSlot
		for rows.Next() {
			var slot models.AvailableSlot
			if err := rows.Scan(&slot.SlotID, &slot.StartTime, &slot.EndTime); err != nil {
				http.Error(w, "Erro ao ler slots", http.StatusInternalServerError)
				return
			}
			slots = append(slots, slot)
		}

		// Renderizar template
		tmpl := template.Must(template.New("available_slots.html").
			Funcs(funcMap).
			ParseFiles("view/partials/available_slots.html"))

		tmpl.Execute(w, map[string]interface{}{
			"Slots": slots,
			"Date":  date,
		})
	}
}

func renderNoSlots(w http.ResponseWriter, date time.Time) {
	tmpl := template.Must(template.New("available_slots.html").
		Funcs(funcMap).
		ParseFiles("view/partials/available_slots.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Slots": nil,
		"Date":  date,
	})
}

// BookAppointment processa o agendamento de uma supervisão
func BookAppointment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Iniciando processo de agendamento...")
		log.Printf("Form values: %+v", r.Form)
		log.Printf("Raw body: %s", r.Body)

		// Obter ID do usuário do contexto
		userID := r.Context().Value(auth.UserIDKey).(int)
		log.Printf("UserID do contexto: %d", userID)

		// Obter ID do slot
		r.ParseForm()
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
			<div class="alert alert-success bg-success text-light border-0">
				<i class="fas fa-check-circle me-2"></i>
				Agendamento realizado com sucesso! Aguardando aprovação do supervisor.
			</div>
		`))
	}
}
