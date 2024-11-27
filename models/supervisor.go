// superviso/models/supervisor.go
package models

type Supervisor struct {
	ID              int     `json:"id"`
	UserID          int     `json:"user_id"`           // Referência ao ID do usuário
	Qualifications  string  `json:"qualifications"`    // Qualificações do supervisor
	PricePerSession float64 `json:"price_per_session"` // Preço por sessão
	Availability    string  `json:"availability"`      // Disponibilidade
	User            *User   `json:"user,omitempty"`    // Dados do usuário relacionados (opcional)
}
