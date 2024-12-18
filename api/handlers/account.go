package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"superviso/api/auth"
)

func CheckAccountAge(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		var createdAt time.Time
		err := db.QueryRow(`
			SELECT created_at 
			FROM users 
			WHERE id = $1`, userID).Scan(&createdAt)

		if err != nil {
			http.Error(w, "Erro ao verificar conta", http.StatusInternalServerError)
			return
		}

		accountAge := time.Since(createdAt)
		isOldEnough := accountAge.Hours() >= 48 // 2 dias

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{
			"isOldEnough": isOldEnough,
		})
	}
}
