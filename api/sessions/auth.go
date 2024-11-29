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
			log.Printf("Erro ao recuperar a sessão: %v\n", err)
			http.Error(w, "Erro ao recuperar sessão", http.StatusInternalServerError)
			return
		}

		log.Printf("Session Values: %+v\n", session.Values)

		if session.Values["user_id"] == nil {
			log.Println("Usuário não autenticado. Redirecionando para login.")
			http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
			return
		}

		log.Printf("Usuário autenticado com ID: %v\n", session.Values["user_id"])
		next.ServeHTTP(w, r)
	})
}
