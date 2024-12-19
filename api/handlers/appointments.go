package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"superviso/api/auth"
	tmpl "superviso/api/template"
	"superviso/websocket"
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

func AcceptAppointmentHandler(db *sql.DB, hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Erro ao processar requisição", http.StatusBadRequest)
			return
		}
		appointmentID := r.FormValue("id")
		appointmentIDInt, err := strconv.Atoi(appointmentID)
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		if appointmentID == "" {
			http.Error(w, "ID do agendamento não fornecido", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value(auth.UserIDKey).(int)

		var (
			supervisorID int
			superviseeID int
		)
		err = db.QueryRow(`
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
		var slotID int
		err = tx.QueryRow(`
			SELECT slot_id 
			FROM appointments 
			WHERE id = $1`, appointmentID).Scan(&slotID)
		if err != nil {
			log.Printf("Erro ao buscar slot_id: %v", err)
			return
		}

		_, err = tx.Exec(`
			UPDATE available_slots 
			SET status = 'booked' 
			WHERE id = (SELECT slot_id FROM appointments WHERE id = $1)`, appointmentID)

		if err != nil {
			http.Error(w, "Erro ao atualizar slot", http.StatusInternalServerError)
			return
		}

		// Notificar todos sobre a atualização do slot
		var appointmentDate, startTime string
		err = tx.QueryRow(`
			SELECT 
				TO_CHAR(s.slot_date, 'DD/MM/YYYY') as formatted_date,
				TO_CHAR(s.start_time, 'HH24:MI') as formatted_time
			FROM appointments a
			JOIN available_slots s ON s.id = a.slot_id
			WHERE a.id = $1`, appointmentID).Scan(&appointmentDate, &startTime)
		if err != nil {
			log.Printf("Erro ao buscar detalhes do slot: %v", err)
			return
		}

		hub.Broadcast(websocket.SlotUpdateMessage{
			Type:         websocket.MessageTypeSlotUpdate,
			SlotID:       slotID,
			Status:       "booked",
			SupervisorID: supervisorID,
			Date:         appointmentDate,
			StartTime:    startTime,
		})

		// Notificar sobre a atualização do agendamento
		hub.Broadcast(websocket.AppointmentUpdateMessage{
			Type:          websocket.MessageTypeAppointmentUpdate,
			AppointmentID: appointmentIDInt,
			Status:        "confirmed",
			Message:       "Agendamento confirmado",
		})

		// Buscar informações para notificação
		var superviseeEmail, supervisorName, superviseeName string
		err = tx.QueryRow(`
			SELECT 
				CONCAT(u.first_name, ' ', u.last_name) as supervisor_name,
				(SELECT email FROM users WHERE id = $2) as supervisee_email,
				(SELECT CONCAT(first_name, ' ', last_name) FROM users WHERE id = $2) as supervisee_name
			FROM appointments a
			JOIN users u ON u.id = a.supervisor_id 
			JOIN available_slots s ON s.id = a.slot_id
			WHERE a.id = $1`, appointmentID, superviseeID).Scan(
			&supervisorName,
			&superviseeEmail, &superviseeName)

		if err != nil {
			log.Printf("Erro ao buscar detalhes do agendamento: %v", err)
			return
		}

		// Criar e enviar notificação
		notificationMsg := fmt.Sprintf("Seu agendamento com %s para %s às %s foi confirmado.",
			supervisorName, appointmentDate, startTime)

		// Notificação para o supervisee
		err = CreateNotificationWithEmail(db, tx, NotificationData{
			UserID:       superviseeID,
			Type:         "appointment_accepted",
			Title:        "Agendamento Confirmado",
			Message:      notificationMsg,
			EmailSubject: "Agendamento Confirmado - Superviso",
			EmailBody: fmt.Sprintf(`
				<img src="static/assets/email/logo.png" alt="Superviso" style="width: 200px; margin-bottom: 20px;">
				<h2>Confirmação</h2>
				<p>Olá %s,</p>
				<p>Você recebeu uma confirmação de agendamento.</p>
				<p>Supervisor: <strong>%s</strong></p>
				<p>Data: <strong>%s</strong></p>
				<p>Horário: <strong>%s</strong></p>
				<p>%s</p>
				<p>Atenciosamente,<br>Equipe Superviso</p>
			`, superviseeName, supervisorName, appointmentDate, startTime, notificationMsg),
		}, hub)

		if err != nil {
			log.Printf("Erro ao criar notificação para supervisee: %v", err)
			return
		}

		// Notificação para o supervisor
		supervisorMsg := fmt.Sprintf("Você confirmou o agendamento com %s para %s às %s.",
			superviseeName, appointmentDate, startTime)

		err = CreateNotificationWithEmail(db, tx, NotificationData{
			UserID:       supervisorID,
			Type:         "appointment_confirmed",
			Title:        "Agendamento Confirmado",
			Message:      supervisorMsg,
			EmailSubject: "Agendamento Confirmado - Superviso",
			EmailBody: fmt.Sprintf(`
				<h2>Agendamento Confirmado</h2>
				<p>Olá,</p>
				<p>%s</p>
				<p>Atenciosamente,<br>Equipe Superviso</p>
			`, supervisorMsg),
		}, hub)

		if err != nil {
			log.Printf("Erro ao criar notificação para supervisor: %v", err)
			return
		}

		// Commit da transação
		if err = tx.Commit(); err != nil {
			http.Error(w, "Erro ao confirmar operação", http.StatusInternalServerError)
			return
		}

		// Retornar a lista atualizada de agendamentos
		AppointmentsHandler(db).ServeHTTP(w, r)
	}
}

func RejectAppointmentHandler(db *sql.DB, hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appointmentID := r.URL.Query().Get("id")
		appointmentIDInt, err := strconv.Atoi(appointmentID)
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		// Verificar se o usuário é o supervisor correto
		var supervisorID int
		err = db.QueryRow(`
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
		var slotID int
		err = tx.QueryRow(`
			SELECT slot_id 
			FROM appointments 
			WHERE id = $1`, appointmentID).Scan(&slotID)
		if err != nil {
			log.Printf("Erro ao buscar slot_id: %v", err)
			return
		}

		_, err = tx.Exec(`
			UPDATE available_slots 
			SET status = 'available' 
			WHERE id = (SELECT slot_id FROM appointments WHERE id = $1)`, appointmentID)

		if err != nil {
			http.Error(w, "Erro ao atualizar slot", http.StatusInternalServerError)
			return
		}

		// Notificar todos sobre a atualização do slot
		var appointmentDate, startTime string
		err = tx.QueryRow(`
			SELECT 
				TO_CHAR(s.slot_date, 'DD/MM/YYYY') as formatted_date,
				TO_CHAR(s.start_time, 'HH24:MI') as formatted_time
			FROM appointments a
			JOIN available_slots s ON s.id = a.slot_id
			WHERE a.id = $1`, appointmentID).Scan(&appointmentDate, &startTime)
		if err != nil {
			log.Printf("Erro ao buscar detalhes do slot: %v", err)
			return
		}

		hub.Broadcast(websocket.SlotUpdateMessage{
			Type:         websocket.MessageTypeSlotUpdate,
			SlotID:       slotID,
			Status:       "available",
			SupervisorID: supervisorID,
			Date:         appointmentDate,
			StartTime:    startTime,
		})

		// Notificar sobre a atualização do agendamento
		hub.Broadcast(websocket.AppointmentUpdateMessage{
			Type:          websocket.MessageTypeAppointmentUpdate,
			AppointmentID: appointmentIDInt,
			Status:        "rejected",
			Message:       "Agendamento rejeitado",
		})

		// Após commit da transação, criar notificação
		var superviseeID int
		var supervisorName, superviseeName string
		err = tx.QueryRow(`
			SELECT 
				a.supervisee_id, 
				CONCAT(u.first_name, ' ', u.last_name) as supervisor_name,
				TO_CHAR(s.slot_date, 'DD/MM/YYYY') as formatted_date,
				TO_CHAR(s.start_time, 'HH24:MI') as formatted_time,
				(SELECT CONCAT(first_name, ' ', last_name) FROM users WHERE id = a.supervisee_id) as supervisee_name
			FROM appointments a
			JOIN users u ON u.id = a.supervisor_id 
			JOIN available_slots s ON s.id = a.slot_id
			WHERE a.id = $1`, appointmentID).Scan(
			&superviseeID, &supervisorName, &appointmentDate, &startTime, &superviseeName)
		if err != nil {
			log.Printf("Erro ao buscar detalhes do agendamento: %v", err)
			return
		}

		notificationMsg := fmt.Sprintf("Seu agendamento com %s para %s às %s foi rejeitado.",
			supervisorName, appointmentDate, startTime)

		// Notificação para o supervisee
		err = CreateNotificationWithEmail(db, tx, NotificationData{
			UserID:       superviseeID,
			Type:         "appointment_rejected",
			Title:        "Agendamento Rejeitado",
			Message:      notificationMsg,
			EmailSubject: "Agendamento Rejeitado - Superviso",
			EmailBody: fmt.Sprintf(`
				<img src="static/assets/email/logo.png" alt="Superviso" style="width: 200px; margin-bottom: 20px;">
				<h2>Rejeição</h2>
				<p>Olá %s,</p>
				<p>Seu agendamento foi rejeitado.</p>
				<p>Supervisor: <strong>%s</strong></p>
				<p>Data: <strong>%s</strong></p>
				<p>Horário: <strong>%s</strong></p>
				<p>%s</p>
				<p>Atenciosamente,<br>Equipe Superviso</p>
			`, superviseeName, supervisorName, appointmentDate, startTime, notificationMsg),
		}, hub)

		if err != nil {
			log.Printf("Erro ao criar notificação para supervisee: %v", err)
			return
		}

		// Notificação para o supervisor
		supervisorMsg := fmt.Sprintf("Você rejeitou o agendamento com %s para %s às %s.",
			superviseeName, appointmentDate, startTime)

		err = CreateNotificationWithEmail(db, tx, NotificationData{
			UserID:       supervisorID,
			Type:         "appointment_rejected_by_me",
			Title:        "Agendamento Rejeitado",
			Message:      supervisorMsg,
			EmailSubject: "Agendamento Rejeitado - Superviso",
			EmailBody: fmt.Sprintf(`
				<h2>Agendamento Rejeitado</h2>
				<p>Olá,</p>
				<p>%s</p>
				<p>Atenciosamente,<br>Equipe Superviso</p>
			`, supervisorMsg),
		}, hub)

		if err != nil {
			log.Printf("Erro ao criar notificação para supervisor: %v", err)
			return
		}

		// Commit da transação
		err = tx.Commit()
		if err != nil {
			http.Error(w, "Erro ao confirmar operação", http.StatusInternalServerError)
			return
		}

		// Retornar a lista atualizada de agendamentos
		AppointmentsHandler(db).ServeHTTP(w, r)
	}
}

func BookAppointment(db *sql.DB, hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		var err error

		// Decodificar dados do request
		var data struct {
			SlotID int `json:"slot_id"`
		}
		err = json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, "Dados inválidos", http.StatusBadRequest)
			return
		}

		// Buscar informações do slot e supervisor
		var (
			supervisorID int
			slotDate     string
			startTime    string
			endTime      string
			slotStatus   string
		)
		err = db.QueryRow(`
			SELECT 
				supervisor_id, 
				TO_CHAR(slot_date, 'DD/MM/YYYY') as formatted_date,
				TO_CHAR(start_time, 'HH24:MI') as formatted_time,
				TO_CHAR(end_time, 'HH24:MI') as formatted_end_time,
				status
			FROM available_slots 
			WHERE id = $1`,
			data.SlotID).Scan(&supervisorID, &slotDate, &startTime, &endTime, &slotStatus)

		if slotStatus != "available" {
			http.Error(w, "Este horário não está mais disponível", http.StatusBadRequest)
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
			if err.Error() == "pq: duplicate key value violates unique constraint \"unique_slot_booking\"" {
				http.Error(w, "Este horário já foi reservado", http.StatusBadRequest)
			} else {
				log.Printf("Erro ao criar agendamento: %v", err)
				http.Error(w, "Erro ao criar agendamento", http.StatusInternalServerError)
			}
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

		// Notificar todos os usuários sobre a atualização do slot
		hub.Broadcast(websocket.SlotUpdateMessage{
			Type:         websocket.MessageTypeSlotUpdate,
			SlotID:       data.SlotID,
			Status:       "pending",
			SupervisorID: supervisorID,
			Date:         slotDate,
			StartTime:    startTime,
		})

		// Buscar informações para notificação
		var superviseeName, supervisorEmail, supervisorName string
		err = db.QueryRow(`
			SELECT 
				(SELECT CONCAT(first_name, ' ', last_name) FROM users WHERE id = $1) as supervisee_name,
				(SELECT email FROM users WHERE id = $2) as supervisor_email,
				(SELECT CONCAT(first_name, ' ', last_name) FROM users WHERE id = $2) as supervisor_name
			FROM users 
			WHERE id = $1`, userID, supervisorID).Scan(&superviseeName, &supervisorEmail, &supervisorName)

		if err != nil {
			log.Printf("Erro ao buscar detalhes: %v", err)
			return
		}

		// Criar notificação para o supervisor
		notificationMsg := fmt.Sprintf("Você recebeu uma nova solicitação de supervisão de %s para %s às %s",
			superviseeName, slotDate, startTime)

		err = CreateNotificationWithEmail(db, tx, NotificationData{
			UserID:       supervisorID,
			Type:         "new_appointment",
			Title:        "Nova Solicitação",
			Message:      notificationMsg,
			EmailSubject: "Nova Solicitação de Supervisão - Superviso",
			EmailBody: fmt.Sprintf(`
				<img src="static/assets/email/logo.png" alt="Superviso" style="width: 200px; margin-bottom: 20px;">
				<h2>Nova Solicitação</h2>
				<p>Olá %s,</p>
				<p>Você recebeu uma nova solicitação de supervisão.</p>
				<p>Supervisionando: <strong>%s</strong></p>
				<p>Data: <strong>%s</strong></p>
				<p>Horário: <strong>%s</strong></p>
				<p>Acesse a plataforma para aceitar ou rejeitar esta solicitação.</p>
			`, supervisorName, superviseeName, slotDate, startTime),
		}, hub)
		if err != nil {
			log.Printf("Erro ao criar notificação: %v", err)
			return
		}

		// Commit da transação
		if err := tx.Commit(); err != nil {
			http.Error(w, "Erro ao confirmar operação", http.StatusInternalServerError)
			return
		}

		// Retornar sucesso
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Agendamento criado com sucesso",
			"id":      appointmentID,
		})
	}
}
