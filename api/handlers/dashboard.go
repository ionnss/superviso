package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"superviso/utils"
	"time"
)

type DashboardData struct {
	NextAppointment     *AppointmentInfo `json:"next_appointment"`
	CompletedSessions   int              `json:"completed_sessions"`
	UnreadNotifications int              `json:"unread_notifications"`
	FavoriteSupervisors int              `json:"favorite_supervisors"`
	RecentActivities    []Activity       `json:"recent_activities"`
}

type AppointmentInfo struct {
	Date           time.Time `json:"date"`
	SupervisorName string    `json:"supervisor_name"`
	SuperviseeName string    `json:"supervisee_name"`
	StartTime      string    `json:"start_time"`
}

type Activity struct {
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func DashboardHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		data := DashboardData{}

		// 1. Fetch next appointment
		err := db.QueryRow(`
			SELECT 
				s.slot_date,
				CONCAT(u.first_name, ' ', u.last_name) as supervisor_name,
				CONCAT(u2.first_name, ' ', u2.last_name) as supervisee_name,
				TO_CHAR(s.start_time, 'HH24:MI') as start_time
			FROM appointments a
			JOIN available_slots s ON a.slot_id = s.id
			JOIN users u ON s.supervisor_id = u.id
			JOIN users u2 ON a.supervisee_id = u2.id
			WHERE (a.supervisor_id = $1 OR a.supervisee_id = $1)
			AND a.status = 'confirmed'
			AND s.slot_date >= CURRENT_DATE
			ORDER BY s.slot_date, s.start_time
			LIMIT 1
		`, userID).Scan(&data.NextAppointment.Date, &data.NextAppointment.SupervisorName,
			&data.NextAppointment.SuperviseeName, &data.NextAppointment.StartTime)

		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "Error fetching next appointment", http.StatusInternalServerError)
			return
		}

		// 2. Count completed sessions
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM appointments 
			WHERE (supervisor_id = $1 OR supervisee_id = $1)
			AND status = 'completed'
		`, userID).Scan(&data.CompletedSessions)

		if err != nil {
			http.Error(w, "Error counting completed sessions", http.StatusInternalServerError)
			return
		}

		// 3. Count unread notifications
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM notifications 
			WHERE user_id = $1 AND read = false
		`, userID).Scan(&data.UnreadNotifications)

		if err != nil {
			http.Error(w, "Error counting notifications", http.StatusInternalServerError)
			return
		}

		// 4. Count favorite supervisors
		err = db.QueryRow(`
			SELECT COUNT(*) 
			FROM favorite_supervisors 
			WHERE supervisee_id = $1
		`, userID).Scan(&data.FavoriteSupervisors)

		if err != nil && err != sql.ErrNoRows {
			http.Error(w, "Error counting favorites", http.StatusInternalServerError)
			return
		}

		// 5. Fetch recent activities
		rows, err := db.Query(`
			SELECT 
				'appointment' as type,
				CASE 
					WHEN a.status = 'confirmed' THEN 'Supervisão confirmada'
					WHEN a.status = 'completed' THEN 'Supervisão realizada'
					WHEN a.status = 'cancelled' THEN 'Supervisão cancelada'
					ELSE 'Supervisão agendada'
				END as message,
				a.updated_at as timestamp
			FROM appointments a
			WHERE (a.supervisor_id = $1 OR a.supervisee_id = $1)
			AND a.updated_at >= NOW() - INTERVAL '30 days'
			ORDER BY a.updated_at DESC
			LIMIT 5
		`, userID)

		if err != nil {
			http.Error(w, "Error fetching activities", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var activity Activity
			err := rows.Scan(&activity.Type, &activity.Message, &activity.Timestamp)
			if err != nil {
				continue
			}
			data.RecentActivities = append(data.RecentActivities, activity)
		}

		// Render the dashboard template with the data
		tmpl := template.Must(template.New("dashboard.html").
			Funcs(template.FuncMap{
				"formatDate": utils.FormatDate,
				"formatTime": utils.FormatTime,
			}).
			ParseFiles("view/dashboard.html"))

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}
