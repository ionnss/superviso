package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"superviso/models"
	"time"

	"superviso/api/auth"
	"superviso/api/email"

	"github.com/gorilla/mux"
)

var notificationTemplates = template.Must(template.New("notifications").
	Funcs(template.FuncMap{
		"formatTimeAgo": formatTimeAgo,
		"formatDateTime": func(t time.Time) string {
			return t.Format("02/01/2006 às 15:04")
		},
	}).
	ParseFiles("view/partials/notifications_list.html"))

func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "agora"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return fmt.Sprintf("há %d min", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("há %d h", hours)
	case diff < 48*time.Hour:
		return "ontem"
	default:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("há %d dias", days)
	}
}

func GetUnreadCountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		notifications, err := models.GetUnreadNotifications(db, userID)
		if err != nil {
			http.Error(w, "Erro ao buscar notificações", http.StatusInternalServerError)
			return
		}

		count := len(notifications)
		if count > 0 {
			w.Write([]byte(fmt.Sprintf("%d", count)))
		}
	}
}

func GetNotificationsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		notifications, err := models.GetUnreadNotifications(db, userID)
		if err != nil {
			http.Error(w, "Erro ao buscar notificações", http.StatusInternalServerError)
			return
		}

		// Renderizar partial de notificações
		err = notificationTemplates.ExecuteTemplate(w, "notifications_list.html", map[string]interface{}{
			"Notifications": notifications,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func MarkNotificationAsReadHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)
		vars := mux.Vars(r)
		notificationID := vars["id"]

		// Converter notificationID para int
		id, err := strconv.Atoi(notificationID)
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		// Marcar como lida
		err = models.MarkNotificationAsRead(db, id, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Notificação não encontrada", http.StatusNotFound)
			} else {
				http.Error(w, "Erro ao atualizar notificação", http.StatusInternalServerError)
			}
			return
		}

		// Retornar sucesso
		w.WriteHeader(http.StatusOK)
	}
}

type NotificationData struct {
	UserID       int
	Title        string
	Message      string
	Type         string
	EmailSubject string
	EmailBody    string
}

// CreateNotificationWithEmail cria uma notificação no sistema e envia um email
func CreateNotificationWithEmail(db *sql.DB, tx *sql.Tx, data NotificationData) error {
	// Criar notificação no sistema
	_, err := tx.Exec(`
		INSERT INTO notifications (user_id, title, message, type)
		VALUES ($1, $2, $3, $4)`,
		data.UserID, data.Title, data.Message, data.Type)

	if err != nil {
		return fmt.Errorf("erro ao criar notificação: %v", err)
	}

	// Buscar email do usuário
	var userEmail string
	err = tx.QueryRow(`
		SELECT email 
		FROM users 
		WHERE id = $1`, data.UserID).Scan(&userEmail)

	if err != nil {
		return fmt.Errorf("erro ao buscar email do usuário: %v", err)
	}

	// Enviar email de forma assíncrona
	go func() {
		err := email.SendEmail(userEmail, data.EmailSubject, data.EmailBody)
		if err != nil {
			log.Printf("Erro ao enviar email: %v", err)
		}
	}()

	return nil
}
