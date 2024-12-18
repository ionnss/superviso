package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"superviso/api/auth"
	"superviso/models"
)

// Variável global para os templates
var templates = template.Must(template.ParseGlob("view/*.html"))

type AppointmentResponse struct {
	ID             int       `json:"id"`
	SupervisorName string    `json:"supervisor_name"`
	SuperviseeName string    `json:"supervisee_name"`
	Date           string    `json:"date"`
	StartTime      string    `json:"start_time"`
	EndTime        string    `json:"end_time"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

func AppointmentsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Obter usuário do contexto
		userID := r.Context().Value(auth.UserIDKey).(int)

		// Verificar se é supervisor
		var isSupervisor bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM supervisor_profiles WHERE user_id = $1)", userID).Scan(&isSupervisor)
		if err != nil {
			http.Error(w, "Erro ao verificar perfil", http.StatusInternalServerError)
			return
		}

		var appointments []AppointmentResponse
		if isSupervisor {
			appointments, err = getPendingSupervisorAppointments(db, userID)
		} else {
			appointments, err = getSuperviseeAppointments(db, userID)
		}

		if err != nil {
			http.Error(w, "Erro ao buscar agendamentos", http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Appointments": appointments,
			"IsSupervisor": isSupervisor,
		}

		err = templates.ExecuteTemplate(w, "appointments.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getPendingSupervisorAppointments(db *sql.DB, userID int) ([]AppointmentResponse, error) {
	query := `
		SELECT 
			a.id,
			CONCAT(u1.first_name, ' ', u1.last_name) as supervisee_name,
			CONCAT(u2.first_name, ' ', u2.last_name) as supervisor_name,
			s.slot_date,
			s.start_time,
			s.end_time,
			a.status,
			a.created_at
		FROM appointments a
		JOIN users u1 ON a.supervisee_id = u1.id
		JOIN users u2 ON a.supervisor_id = u2.id
		JOIN available_slots s ON a.slot_id = s.id
		WHERE a.supervisor_id = $1 AND a.status = 'pending'
		ORDER BY s.slot_date, s.start_time`

	rows, err := db.Query(query, userID)
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

func getSuperviseeAppointments(db *sql.DB, userID int) ([]AppointmentResponse, error) {
	query := `
		SELECT 
			a.id,
			CONCAT(u1.first_name, ' ', u1.last_name) as supervisee_name,
			CONCAT(u2.first_name, ' ', u2.last_name) as supervisor_name,
			s.slot_date,
			s.start_time,
			s.end_time,
			a.status,
			a.created_at
		FROM appointments a
		JOIN users u1 ON a.supervisee_id = u1.id
		JOIN users u2 ON a.supervisor_id = u2.id
		JOIN available_slots s ON a.slot_id = s.id
		WHERE a.supervisee_id = $1
		ORDER BY s.slot_date, s.start_time`

	rows, err := db.Query(query, userID)
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

		err = tx.Commit()
		if err != nil {
			http.Error(w, "Erro ao confirmar operação", http.StatusInternalServerError)
			return
		}

		// Após commit da transação, criar notificação
		// Buscar informações do agendamento
		var superviseeID int
		var supervisorName, appointmentDate, startTime string
		err = db.QueryRow(`
			SELECT 
				a.supervisee_id, 
				CONCAT(u.first_name, ' ', u.last_name) as supervisor_name,
				s.slot_date,
				s.start_time
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
			Type:   "appointment_accepted",
			Title:  "Agendamento Confirmado",
			Message: fmt.Sprintf("Seu agendamento com %s para %s às %s foi confirmado.",
				supervisorName, appointmentDate, startTime),
		}

		err = models.CreateNotification(db, notification)
		if err != nil {
			log.Printf("Erro ao criar notificação: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func RejectAppointmentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appointmentID := r.URL.Query().Get("id")
		userID := r.Context().Value("user_id").(int)

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

		w.WriteHeader(http.StatusOK)
	}
}
