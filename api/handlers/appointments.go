package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"superviso/api/auth"
	"superviso/api/email"
	tmpl "superviso/api/template"
	"superviso/models"
)

type AppointmentResponse struct {
	ID             int       `json:"id"`
	SupervisorName string    `json:"supervisor_name"`
	SuperviseeName string    `json:"supervisee_name"`
	Date           time.Time `json:"date"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

func AppointmentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		var isSupervisor bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM supervisor_profiles WHERE user_id = $1)", userID).Scan(&isSupervisor)
		if err != nil {
			http.Error(w, "Erro ao verificar perfil", http.StatusInternalServerError)
			return
		}

		var pendingAppointments, confirmedAppointments, historicAppointments []AppointmentResponse
		if isSupervisor {
			pendingAppointments, err = getSupervisorAppointmentsByStatus(db, userID, "pending")
			if err != nil {
				http.Error(w, "Erro ao buscar agendamentos pendentes", http.StatusInternalServerError)
				return
			}
			confirmedAppointments, err = getSupervisorAppointmentsByStatus(db, userID, "confirmed")
			if err != nil {
				http.Error(w, "Erro ao buscar agendamentos confirmados", http.StatusInternalServerError)
				return
			}
			historicAppointments, err = getSupervisorAppointmentsByStatus(db, userID, "rejected")
			if err != nil {
				http.Error(w, "Erro ao buscar histórico de agendamentos", http.StatusInternalServerError)
				return
			}
		} else {
			pendingAppointments, err = getSuperviseeAppointmentsByStatus(db, userID, "pending")
			if err != nil {
				http.Error(w, "Erro ao buscar agendamentos pendentes", http.StatusInternalServerError)
				return
			}
			confirmedAppointments, err = getSuperviseeAppointmentsByStatus(db, userID, "confirmed")
			if err != nil {
				http.Error(w, "Erro ao buscar agendamentos confirmados", http.StatusInternalServerError)
				return
			}
			historicAppointments, err = getSuperviseeAppointmentsByStatus(db, userID, "rejected")
			if err != nil {
				http.Error(w, "Erro ao buscar histórico de agendamentos", http.StatusInternalServerError)
				return
			}
		}

		// Criar template com as funções necessárias
		tmpl := template.Must(template.New("appointments.html").
			Funcs(tmpl.TemplateFuncs).
			ParseFiles("view/appointments.html"))

		// Executar template com os dados
		err = tmpl.Execute(w, map[string]interface{}{
			"PendingAppointments":   pendingAppointments,
			"ConfirmedAppointments": confirmedAppointments,
			"HistoricAppointments":  historicAppointments,
			"IsSupervisor":          isSupervisor,
		})

		if err != nil {
			log.Printf("Erro ao renderizar template: %v", err)
			http.Error(w, "Erro ao renderizar página", http.StatusInternalServerError)
			return
		}
	}
}

func getSupervisorAppointmentsByStatus(db *sql.DB, userID int, status string) ([]AppointmentResponse, error) {
	query := `
		SELECT 
			a.id,
			CONCAT(u1.first_name, ' ', u1.last_name) as supervisee_name,
			CONCAT(u2.first_name, ' ', u2.last_name) as supervisor_name,
			s.slot_date::timestamp,
			(s.slot_date + s.start_time)::timestamp,
			(s.slot_date + s.end_time)::timestamp,
			a.status,
			a.created_at
		FROM appointments a
		JOIN users u1 ON a.supervisee_id = u1.id
		JOIN users u2 ON a.supervisor_id = u2.id
		JOIN available_slots s ON a.slot_id = s.id
		WHERE a.supervisor_id = $1 AND a.status = $2
		ORDER BY s.slot_date, s.start_time`

	rows, err := db.Query(query, userID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []AppointmentResponse
	for rows.Next() {
		var apt AppointmentResponse
		err := rows.Scan(
			&apt.ID,
			&apt.SuperviseeName,
			&apt.SupervisorName,
			&apt.Date,
			&apt.StartTime,
			&apt.EndTime,
			&apt.Status,
			&apt.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		appointments = append(appointments, apt)
	}

	return appointments, nil
}

func getSuperviseeAppointmentsByStatus(db *sql.DB, userID int, status string) ([]AppointmentResponse, error) {
	query := `
		SELECT 
			a.id,
			CONCAT(u1.first_name, ' ', u1.last_name) as supervisee_name,
			CONCAT(u2.first_name, ' ', u2.last_name) as supervisor_name,
			s.slot_date::timestamp,
			(s.slot_date + s.start_time)::timestamp,
			(s.slot_date + s.end_time)::timestamp,
			a.status,
			a.created_at
		FROM appointments a
		JOIN users u1 ON a.supervisee_id = u1.id
		JOIN users u2 ON a.supervisor_id = u2.id
		JOIN available_slots s ON a.slot_id = s.id
		WHERE a.supervisee_id = $1 AND a.status = $2
		ORDER BY s.slot_date, s.start_time`

	rows, err := db.Query(query, userID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []AppointmentResponse
	for rows.Next() {
		var apt AppointmentResponse
		err := rows.Scan(
			&apt.ID,
			&apt.SuperviseeName,
			&apt.SupervisorName,
			&apt.Date,
			&apt.StartTime,
			&apt.EndTime,
			&apt.Status,
			&apt.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		appointments = append(appointments, apt)
	}

	return appointments, nil
}

func AcceptAppointmentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appointmentID := r.URL.Query().Get("id")
		userID := r.Context().Value(auth.UserIDKey).(int)

		var (
			supervisorID int
			superviseeID int
		)
		err := db.QueryRow(`
			SELECT supervisor_id, supervisee_id
			FROM appointments 
			WHERE id = $1`, appointmentID).Scan(&supervisorID, &superviseeID)

		if err != nil {
			http.Error(w, "Agendamento não encontrado", http.StatusNotFound)
			return
		}

		if supervisorID != userID {
			http.Error(w, "Não autorizado", http.StatusForbidden)
			return
		}

		// Iniciar transação
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Atualizar status do agendamento
		_, err = tx.Exec(`
			UPDATE appointments 
			SET status = 'confirmed', updated_at = NOW() 
			WHERE id = $1`, appointmentID)

		if err != nil {
			http.Error(w, "Erro ao atualizar agendamento", http.StatusInternalServerError)
			return
		}

		// Atualizar slot para booked
		_, err = tx.Exec(`
			UPDATE available_slots 
			SET status = 'booked' 
			WHERE id = (SELECT slot_id FROM appointments WHERE id = $1)`, appointmentID)

		if err != nil {
			http.Error(w, "Erro ao atualizar slot", http.StatusInternalServerError)
			return
		}

		// Commit da transação
		if err = tx.Commit(); err != nil {
			http.Error(w, "Erro ao confirmar operação", http.StatusInternalServerError)
			return
		}

		// Buscar informações para notificação
		var supervisorName, appointmentDate, startTime, superviseeEmail string
		err = db.QueryRow(`
			SELECT 
				CONCAT(u.first_name, ' ', u.last_name) as supervisor_name,
				TO_CHAR(s.slot_date, 'DD/MM/YYYY') as formatted_date,
				TO_CHAR(s.start_time, 'HH24:MI') as formatted_time,
				(SELECT email FROM users WHERE id = $2) as supervisee_email
			FROM appointments a
			JOIN users u ON u.id = a.supervisor_id 
			JOIN available_slots s ON s.id = a.slot_id
			WHERE a.id = $1`, appointmentID, superviseeID).Scan(
			&supervisorName, &appointmentDate, &startTime, &superviseeEmail)

		if err != nil {
			log.Printf("Erro ao buscar detalhes do agendamento: %v", err)
			return
		}

		// Criar e enviar notificação
		notificationMsg := fmt.Sprintf("Seu agendamento com %s para %s às %s foi confirmado.",
			supervisorName, appointmentDate, startTime)

		notification := &models.Notification{
			UserID:  superviseeID,
			Type:    "appointment_accepted",
			Title:   "Agendamento Confirmado",
			Message: notificationMsg,
		}

		if err = models.CreateNotification(db, notification); err != nil {
			log.Printf("Erro ao criar notificação: %v", err)
			return
		}

		// Enviar email de forma assíncrona
		go func() {
			err := email.SendEmail(
				superviseeEmail,
				"Agendamento Confirmado - Superviso",
				fmt.Sprintf(`
					<h2>Agendamento Confirmado</h2>
					<p>Olá,</p>
					<p>%s</p>
					<p>Atenciosamente,<br>Equipe Superviso</p>
				`, notificationMsg),
			)
			if err != nil {
				log.Printf("Erro ao enviar email: %v", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
	}
}

func RejectAppointmentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appointmentID := r.URL.Query().Get("id")

		// Verificar se o usuário é o supervisor correto
		var supervisorID int
		err := db.QueryRow(`
				SELECT supervisor_id 
				FROM appointments 
				WHERE id = $1`, appointmentID).Scan(&supervisorID)

		if err != nil {
			http.Error(w, "Agendamento não encontrado", http.StatusNotFound)
			return
		}

		// Iniciar transação
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Atualizar status do agendamento
		_, err = tx.Exec(`
			UPDATE appointments 
			SET status = 'rejected', updated_at = NOW() 
			WHERE id = $1`, appointmentID)

		if err != nil {
			http.Error(w, "Erro ao atualizar agendamento", http.StatusInternalServerError)
			return
		}

		// Liberar o slot
		_, err = tx.Exec(`
			UPDATE available_slots 
			SET status = 'available' 
			WHERE id = (SELECT slot_id FROM appointments WHERE id = $1)`, appointmentID)

		if err != nil {
			http.Error(w, "Erro ao atualizar slot", http.StatusInternalServerError)
			return
		}

		err = tx.Commit()
		if err != nil {
			http.Error(w, "Erro ao confirmar operação", http.StatusInternalServerError)
			return
		}

		// Após commit da transação, criar notificação
		var superviseeID int
		var supervisorName, appointmentDate, startTime string
		err = db.QueryRow(`
			SELECT 
				a.supervisee_id, 
				CONCAT(u.first_name, ' ', u.last_name) as supervisor_name,
				TO_CHAR(s.slot_date, 'DD/MM/YYYY') as formatted_date,
				TO_CHAR(s.start_time, 'HH24:MI') as formatted_time
			FROM appointments a
			JOIN users u ON u.id = a.supervisor_id 
			JOIN available_slots s ON s.id = a.slot_id
			WHERE a.id = $1`, appointmentID).Scan(&superviseeID, &supervisorName, &appointmentDate, &startTime)
		if err != nil {
			log.Printf("Erro ao buscar detalhes do agendamento: %v", err)
			return
		}

		notification := &models.Notification{
			UserID: superviseeID,
			Type:   "appointment_rejected",
			Title:  "Agendamento Rejeitado",
			Message: fmt.Sprintf("Seu agendamento com %s para %s às %s foi rejeitado.",
				supervisorName, appointmentDate, startTime),
		}

		err = models.CreateNotification(db, notification)
		if err != nil {
			log.Printf("Erro ao criar notificação: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func CreateAppointmentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)
		slotID := r.FormValue("slot_id")

		// Iniciar transação
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Verificar se o slot está disponível
		var supervisorID int
		var slotDate string
		var startTime string
		err = tx.QueryRow(`
			SELECT supervisor_id, slot_date, start_time 
			FROM available_slots 
			WHERE id = $1 AND status = 'available'`,
			slotID).Scan(&supervisorID, &slotDate, &startTime)
		if err != nil {
			http.Error(w, "Horário não disponível", http.StatusBadRequest)
			return
		}

		// Criar o agendamento
		var appointmentID int
		err = tx.QueryRow(`
			INSERT INTO appointments (supervisor_id, supervisee_id, slot_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, 'pending', NOW(), NOW())
			RETURNING id`,
			supervisorID, userID, slotID).Scan(&appointmentID)
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
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Erro ao confirmar operação", http.StatusInternalServerError)
			return
		}

		// Após commit, criar notificação para o supervisor
		var superviseeName string
		err = db.QueryRow(`
			SELECT CONCAT(first_name, ' ', last_name)
			FROM users
			WHERE id = $1`,
			userID).Scan(&superviseeName)
		if err != nil {
			log.Printf("Erro ao buscar nome do supervisionando: %v", err)
			return
		}

		notification := &models.Notification{
			UserID: supervisorID,
			Type:   "new_appointment_request",
			Title:  "Nova Solicitação de Agendamento",
			Message: fmt.Sprintf("%s solicitou um agendamento para %s às %s.",
				superviseeName, slotDate, startTime),
		}

		err = models.CreateNotification(db, notification)
		if err != nil {
			log.Printf("Erro ao criar notificação: %v", err)
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func BookAppointment(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		// Decodificar dados do request
		var data struct {
			SlotID int `json:"slot_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Dados inválidos", http.StatusBadRequest)
			return
		}

		// Buscar informações do slot e supervisor
		var (
			supervisorID int
			slotDate     string
			startTime    string
			endTime      string
		)
		err := db.QueryRow(`
			SELECT 
				supervisor_id, 
				TO_CHAR(slot_date, 'DD/MM/YYYY') as formatted_date,
				TO_CHAR(start_time, 'HH24:MI') as formatted_time,
				TO_CHAR(end_time, 'HH24:MI') as formatted_end_time
			FROM available_slots 
			WHERE id = $1 AND status = 'available'`,
			data.SlotID).Scan(&supervisorID, &slotDate, &startTime, &endTime)

		if err != nil {
			http.Error(w, "Slot não disponível", http.StatusBadRequest)
			return
		}

		// Iniciar transação
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		// Criar o agendamento
		var appointmentID int
		err = tx.QueryRow(`
			INSERT INTO appointments (supervisor_id, supervisee_id, slot_id, status)
			VALUES ($1, $2, $3, 'pending')
			RETURNING id`,
			supervisorID, userID, data.SlotID).Scan(&appointmentID)

		if err != nil {
			log.Printf("Erro ao criar agendamento: %v", err)
			http.Error(w, "Erro ao criar agendamento", http.StatusInternalServerError)
			return
		}

		// Atualizar status do slot
		_, err = tx.Exec(`
			UPDATE available_slots 
			SET status = 'pending' 
			WHERE id = $1`, data.SlotID)

		if err != nil {
			log.Printf("Erro ao atualizar slot: %v", err)
			http.Error(w, "Erro ao atualizar disponibilidade", http.StatusInternalServerError)
			return
		}

		// Buscar nome do supervisionando para a notificação
		var superviseeName string
		var supervisorEmail string
		err = tx.QueryRow(`
			SELECT CONCAT(first_name, ' ', last_name), email
			FROM users
			WHERE id = $1`, supervisorID).Scan(&superviseeName, &supervisorEmail)

		if err != nil {
			log.Printf("Erro ao buscar nome do supervisionando: %v", err)
			http.Error(w, "Erro ao criar notificação", http.StatusInternalServerError)
			return
		}

		// Criar notificação para o supervisor
		notificationMsg := fmt.Sprintf("Você recebeu uma nova solicitação de supervisão de %s para %s às %s",
			superviseeName,
			slotDate,
			startTime)

		// Criar notificação no sistema
		notification := &models.Notification{
			UserID:  supervisorID,
			Type:    "appointment_request",
			Title:   "Nova Solicitação de Supervisão",
			Message: notificationMsg,
		}

		err = models.CreateNotification(db, notification)
		if err != nil {
			log.Printf("Erro ao criar notificação no sistema: %v", err)
			http.Error(w, "Erro ao criar notificação", http.StatusInternalServerError)
			return
		}

		// Commit da transação
		if err := tx.Commit(); err != nil {
			http.Error(w, "Erro ao confirmar operação", http.StatusInternalServerError)
			return
		}

		// Enviar email para o supervisor
		go func() {
			err := email.SendEmail(
				supervisorEmail,
				"Nova Solicitação de Supervisão - Superviso",
				fmt.Sprintf(`
					<h2>Nova Solicitação de Supervisão</h2>
					<p>Olá,</p>
					<p>%s</p>
					<p>Acesse a plataforma para aceitar ou rejeitar esta solicitação.</p>
					<p>Atenciosamente,<br>Equipe Superviso</p>
				`, notificationMsg),
			)
			if err != nil {
				log.Printf("Erro ao enviar email: %v", err)
			}
		}()

		// Retornar sucesso
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Agendamento criado com sucesso",
			"id":      appointmentID,
		})
	}
}
