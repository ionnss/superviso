// superviso/api/sessions/auth.go
package sessions

import (
	"log"
	"net/http"
)

// Logout encerra a sessão do usuário
func Logout(w http.ResponseWriter, r *http.Request) {
	session, err := GetSession(r)
	if err != nil {
		log.Printf("Erro ao recuperar a sessão no logout: %v\n", err)
		http.Error(w, "Erro ao recuperar a sessão", http.StatusInternalServerError)
		return
	}

	log.Printf("Antes de limpar: Session Values: %+v\n", session.Values)

	// Limpa os valores da sessão e invalida o cookie
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1 // Expira imediatamente
	err = session.Save(r, w)
	if err != nil {
		log.Printf("Erro ao salvar a sessão no logout: %v\n", err)
		http.Error(w, "Erro ao encerrar a sessão", http.StatusInternalServerError)
		return
	}

	// Redefine manualmente o cookie para expirar imediatamente
	http.SetCookie(w, &http.Cookie{
		Name:     "superviso-session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	log.Printf("Após limpar: Session Values: %+v\n", session.Values)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logout realizado com sucesso!"))
}

// AuthMiddleware verifica se o usuário está autenticado
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := GetSession(r)
		if err != nil {
			log.Printf("Erro ao recuperar a sessão no middleware: %v\n", err)
			http.Error(w, "Erro ao recuperar sessão", http.StatusInternalServerError)
			return
		}

		log.Printf("Middleware - Session Values: %+v\n", session.Values)
		log.Printf("Middleware - MaxAge: %d\n", session.Options.MaxAge)

		if session.Options.MaxAge == -1 || session.Values["user_id"] == nil {
			log.Println("Middleware - Usuário não autenticado.")
			http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
			return
		}

		log.Printf("Middleware - Usuário autenticado com ID: %v\n", session.Values["user_id"])
		next.ServeHTTP(w, r)
	})
}
