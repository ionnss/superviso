// superviso/api/sessions/auth.go
package sessions

import (
	"log"
	"net/http"
)

// AuthMiddleware verifica se o usuário está autenticado
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := GetSession(r)
		if err != nil {
			http.Error(w, "Erro ao recuperar sessão", http.StatusInternalServerError)
			return
		}

		log.Printf("Session Values: %+v\n", session.Values)
		if session.Values["user_id"] == nil {
			http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
