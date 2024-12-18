package models

import (
	"database/sql"
	"time"
)

type Notification struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateNotification(db *sql.DB, n *Notification) error {
	_, err := db.Exec(`
		INSERT INTO notifications (user_id, type, title, message, read, created_at, updated_at)
		VALUES ($1, $2, $3, $4, false, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		n.UserID, n.Type, n.Title, n.Message)
	return err
}

func GetUnreadNotifications(db *sql.DB, userID int) ([]Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, read, created_at, updated_at
		FROM notifications
		WHERE user_id = $1 AND read = FALSE
		ORDER BY created_at DESC`

	return getNotifications(db, query, userID)
}

func MarkNotificationAsRead(db *sql.DB, notificationID, userID int) error {
	query := `
		UPDATE notifications 
		SET read = TRUE, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`

	result, err := db.Exec(query, notificationID, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func getNotifications(db *sql.DB, query string, args ...interface{}) ([]Notification, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Message,
			&n.Read,
			&n.CreatedAt,
			&n.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}
