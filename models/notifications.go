package models

import (
	"database/sql"
)

func GetNotifications(db *sql.DB, userID int) ([]Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, read, created_at, updated_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message,
			&n.Read, &n.CreatedAt, &n.UpdatedAt)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}

	return notifications, nil
}
