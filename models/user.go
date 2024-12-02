// superviso/models/user.go
package models

import "time"

// AuthUser define uma interface comum para ambos os tipos de usuário
type AuthUser interface {
	GetID() int
	GetEmail() string
	GetPasswordHash() string
	GetUserRole() string
	GetFailedLoginAttempts() int
	GetLastFailedLogin() time.Time
}

// Supervisor represents a supervisor user in the system
type Supervisor struct {
	ID                  int       `json:"id"`
	FirstName           string    `json:"first_name"`
	LastName            string    `json:"last_name"`
	CPF                 string    `json:"cpf"`
	Email               string    `json:"email"`
	PasswordHash        string    `json:"-"` // Armazenado no banco
	CRP                 string    `json:"crp"`
	TheoryApproach      string    `json:"theory_approach"`
	Qualifications      string    `json:"qualifications"`
	UserRole            string    `json:"user_role"`
	CreatedAt           time.Time `json:"created_at"`
	FailedLoginAttempts int       `json:"-"`
	LastFailedLogin     time.Time `json:"-"`
}

// Métodos para implementar a interface AuthUser para Supervisor
func (s *Supervisor) GetID() int                    { return s.ID }
func (s *Supervisor) GetEmail() string              { return s.Email }
func (s *Supervisor) GetPasswordHash() string       { return s.PasswordHash }
func (s *Supervisor) GetUserRole() string           { return s.UserRole }
func (s *Supervisor) GetFailedLoginAttempts() int   { return s.FailedLoginAttempts }
func (s *Supervisor) GetLastFailedLogin() time.Time { return s.LastFailedLogin }

// Supervisionated represents a supervised user in the system
type Supervisionated struct {
	ID                  int       `json:"id"`
	FirstName           string    `json:"first_name"`
	LastName            string    `json:"last_name"`
	CPF                 string    `json:"cpf"`
	Email               string    `json:"email"`
	PasswordHash        string    `json:"-"` // Armazenado no banco
	CRP                 string    `json:"crp"`
	TheoryApproach      string    `json:"theory_approach"`
	Qualifications      string    `json:"qualifications"`
	UserRole            string    `json:"user_role"`
	CreatedAt           time.Time `json:"created_at"`
	FailedLoginAttempts int       `json:"-"`
	LastFailedLogin     time.Time `json:"-"`
}

// Métodos para implementar a interface AuthUser para Supervisionated
func (s *Supervisionated) GetID() int                    { return s.ID }
func (s *Supervisionated) GetEmail() string              { return s.Email }
func (s *Supervisionated) GetPasswordHash() string       { return s.PasswordHash }
func (s *Supervisionated) GetUserRole() string           { return s.UserRole }
func (s *Supervisionated) GetFailedLoginAttempts() int   { return s.FailedLoginAttempts }
func (s *Supervisionated) GetLastFailedLogin() time.Time { return s.LastFailedLogin }

// LoginCredentials representa as credenciais de login
type LoginCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SupervisorAvailability represents the availability slots for supervisors
type SupervisorAvailability struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	AvailabilityDay  string    `json:"availability_day"`
	AvailabilityTime time.Time `json:"availability_time"`
	PricePerSession  float64   `json:"price_per_session"`
}
