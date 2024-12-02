// superviso/api/sessions/sessions.go
package sessions

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// Configura o armazenamento de sessões
var store = sessions.NewCookieStore([]byte("super-secret-key"))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 1 day
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
}

// GetSession retorna a sessão do usuário
func GetSession(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, "superviso-session")
}
