package appointment

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
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
			SELECT u.id, u.first_name, u.last_name, u.crp
			FROM users u
			JOIN supervisor_profiles sp ON u.id = sp.user_id
			WHERE u.id = $1`,
			supervisorID).Scan(
			&supervisor.UserID,
			&supervisor.FirstName,
			&supervisor.LastName,
			&supervisor.CRP,
		)
		if err != nil {
			http.Error(w, "Erro ao buscar supervisor", http.StatusInternalServerError)
			return
		}

		// Buscar horários semanais do supervisor
		rows, err := db.Query(`
			SELECT weekday, start_time, end_time 
			FROM supervisor_weekly_hours 
			WHERE supervisor_id = $1 
			ORDER BY weekday`,
			supervisorID)
		if err != nil {
			http.Error(w, "Erro ao buscar horários", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var weeklyHours []models.SupervisorWeeklyHours
		for rows.Next() {
			var h models.SupervisorWeeklyHours
			err := rows.Scan(&h.Weekday, &h.StartTime, &h.EndTime)
			if err != nil {
				http.Error(w, "Erro ao ler horários", http.StatusInternalServerError)
				return
			}
			weeklyHours = append(weeklyHours, h)
		}

		// Preparar dados para o template
		data := struct {
			Supervisor  models.Supervisor
			WeeklyHours []models.SupervisorWeeklyHours
			WeekDays    []int
		}{
			Supervisor:  supervisor,
			WeeklyHours: weeklyHours,
			WeekDays:    []int{1, 2, 3, 4, 5, 6, 7},
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
		weekday := r.URL.Query().Get("weekday")

		// Buscar horário do supervisor para este dia
		var startTime, endTime string
		err := db.QueryRow(`
			SELECT start_time, end_time 
			FROM supervisor_weekly_hours 
			WHERE supervisor_id = $1 AND weekday = $2`,
			supervisorID, weekday).Scan(&startTime, &endTime)
		if err == sql.ErrNoRows {
			// Retornar array vazio se não houver horário configurado
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]models.AvailableSlot{})
			return
		}
		if err != nil {
			http.Error(w, "Erro ao buscar horários", http.StatusInternalServerError)
			return
		}

		// Verificar períodos de disponibilidade
		var isAvailable bool
		err = db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM supervisor_availability_periods
				WHERE supervisor_id = $1 
				AND CURRENT_DATE BETWEEN start_date AND end_date
			)`, supervisorID).Scan(&isAvailable)
		if err != nil {
			http.Error(w, "Erro ao verificar disponibilidade", http.StatusInternalServerError)
			return
		}

		if !isAvailable {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]models.AvailableSlot{})
			return
		}

		// Gerar slots para as próximas 4 semanas
		supID, err := strconv.Atoi(supervisorID)
		if err != nil {
			http.Error(w, "ID do supervisor inválido", http.StatusBadRequest)
			return
		}
		slots := generateSlots(supID, weekday, startTime, endTime)

		// Filtrar slots já agendados
		availableSlots := filterAvailableSlots(db, slots)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(availableSlots)
	}
}

// Funções auxiliares
func formatWeekday(day int) string {
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
			<div class="alert alert-success">
				<i class="fas fa-check-circle me-2"></i>
				Agendamento realizado com sucesso! Aguardando confirmação do supervisor.
			</div>
		`))
	}
}

// GenerateSlots gera slots de horário para as próximas 4 semanas
func generateSlots(supervisorID int, weekday string, startTime, endTime string) []models.AvailableSlot {
	var slots []models.AvailableSlot
	now := time.Now()

	// Gerar slots para as próximas 4 semanas
	for week := 0; week < 4; week++ {
		date := nextWeekday(now.AddDate(0, 0, week*7), weekday)

		slot := models.AvailableSlot{
			SupervisorID: supervisorID,
			SlotDate:     date,
			StartTime:    startTime,
			EndTime:      endTime,
			Status:       "available",
		}
		slots = append(slots, slot)
	}
	return slots
}

// nextWeekday retorna a próxima data para um dia específico da semana
func nextWeekday(start time.Time, weekday string) time.Time {
	// Converter string para int
	day, err := strconv.Atoi(weekday)
	if err != nil {
		return start
	}

	// Ajustar para o formato time.Weekday (0-6, onde 0 é Domingo)
	// Nossa aplicação usa 1-7, onde 1 é Segunda
	if day == 7 {
		day = 0
	}

	date := start
	for date.Weekday() != time.Weekday(day) {
		date = date.AddDate(0, 0, 1)
	}
	return date
}

// filterAvailableSlots filtra os slots que já estão agendados
func filterAvailableSlots(db *sql.DB, slots []models.AvailableSlot) []models.AvailableSlot {
	var availableSlots []models.AvailableSlot
	for _, slot := range slots {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM available_slots 
				WHERE supervisor_id = $1 
				AND slot_date = $2 
				AND start_time = $3
				AND status = 'booked'
			)`,
			slot.SupervisorID, slot.SlotDate, slot.StartTime).Scan(&exists)

		if err != nil {
			log.Printf("Erro ao verificar slot %v: %v", slot, err)
			continue
		}
		if !exists {
			availableSlots = append(availableSlots, slot)
		}
	}
	return availableSlots
}
