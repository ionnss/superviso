Vou criar a documentação completa agora. Vou dividir em várias mensagens para facilitar a leitura e revisão, mas será um documento único e completo.

Primeira parte do markdown:

````markdown
# Superviso - Documentação Completa de Refatoração

## Índice
1. [Visão Geral](#visão-geral)
2. [Estrutura do Projeto](#estrutura-do-projeto)
3. [Banco de Dados](#banco-de-dados)
4. [Regras de Negócio](#regras-de-negócio)
5. [Implementação Técnica](#implementação-técnica)
6. [Fluxos de Usuário](#fluxos-de-usuário)
7. [Segurança](#segurança)
8. [Testes](#testes)

## Visão Geral

### Objetivo
Sistema de agendamento de supervisões que permite:
- Supervisores disponibilizarem horários
- Supervisionandos agendarem horários
- Gerenciamento de agendamentos
- Notificações em tempo real
- Comunicação via email

### Tecnologias
- Backend: Go 1.21+
- Frontend: HTMX + Bootstrap
- Banco de Dados: PostgreSQL
- Real-time: WebSocket
- Email: SMTP

## Estrutura do Projeto
```
/superviso
├── main.go                 # Ponto de entrada da aplicação
├── handlers/              # Handlers HTTP
│   ├── auth.go           # Autenticação e autorização
│   ├── appointments.go   # Gestão de agendamentos
│   ├── notifications.go  # Sistema de notificações
│   └── slots.go         # Gestão de horários
├── templates/            # Templates HTML
│   ├── layout/
│   │   └── base.html    # Template base
│   ├── auth/
│   │   ├── login.html
│   │   └── register.html
│   └── appointments/
│       ├── list.html
│       ├── new.html
│       └── calendar.html
├── static/
│   ├── css/
│   │   └── style.css    # Único arquivo CSS
│   ├── js/
│   │   └── app.js       # Único arquivo JavaScript
│   └── assets/
│       └── images/
└── db/
    ├── migrations/      # Arquivos SQL de migração
    └── queries/         # Queries SQL comuns
```

## Banco de Dados

### Esquema
```sql
-- Usuários
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    is_supervisor BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Agendamentos
CREATE TABLE appointments (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    supervisee_id INT REFERENCES users(id),
    date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    status TEXT CHECK (status IN ('pending', 'confirmed', 'rejected', 'cancelled')),
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT no_overlap EXCLUDE USING gist (
        supervisor_id WITH =,
        daterange(date, date, '[]') WITH &&,
        timerange(start_time, end_time, '[]') WITH &&
    )
);

-- Notificações
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Índices
CREATE INDEX idx_appointments_supervisor ON appointments(supervisor_id);
CREATE INDEX idx_appointments_supervisee ON appointments(supervisee_id);
CREATE INDEX idx_appointments_date ON appointments(date);
CREATE INDEX idx_notifications_user ON notifications(user_id);
```

### Queries Principais
```sql
-- Buscar horários disponíveis
SELECT 
    a.id, 
    a.date, 
    a.start_time, 
    a.end_time,
    u.first_name || ' ' || u.last_name as supervisor_name
FROM appointments a
JOIN users u ON u.id = a.supervisor_id
WHERE a.status = 'available'
AND a.date >= CURRENT_DATE
ORDER BY a.date, a.start_time;

-- Buscar agendamentos pendentes do supervisor
SELECT 
    a.*, 
    u.first_name || ' ' || u.last_name as supervisee_name
FROM appointments a
JOIN users u ON u.id = a.supervisee_id
WHERE a.supervisor_id = $1 
AND a.status = 'pending'
ORDER BY a.date, a.start_time;

-- Buscar notificações não lidas
SELECT *
FROM notifications
WHERE user_id = $1 
AND read = false
ORDER BY created_at DESC;
```
````

Continuando com a próxima parte do markdown:

````markdown
## Regras de Negócio

### Usuários
1. **Tipos de Usuário**
   - Supervisor
   - Supervisionando
   - Não é permitido ser ambos simultaneamente

2. **Registro e Autenticação**
   - Email único
   - Senha com mínimo 8 caracteres
   - Sessão expira em 24 horas

### Agendamentos
1. **Regras de Horários**
   - Duração fixa de 1 hora
   - Não permite sobreposição
   - Mínimo 24h de antecedência
   - Máximo 3 meses de antecedência

2. **Estados do Agendamento**
   ```
   pending -> confirmed/rejected
   confirmed -> cancelled (até 24h antes)
   ```

3. **Validações**
   - Supervisor não pode agendar consigo mesmo
   - Supervisionando não pode ter mais de 1 agendamento pendente com mesmo supervisor
   - Horário deve estar disponível no momento do agendamento

### Notificações
1. **Tipos de Notificação**
   - Nova solicitação
   - Confirmação
   - Rejeição
   - Cancelamento
   - Lembrete (24h antes)

2. **Canais**
   - WebSocket (tempo real)
   - Email (assíncrono)
   - Interface (persistente)

## Implementação Técnica

### Configuração Principal (main.go)
```go
func main() {
    // Carregar variáveis de ambiente
    if err := godotenv.Load(); err != nil {
        log.Fatal("Erro ao carregar .env")
    }

    // Conectar ao banco
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Configurar templates
    tmpl := template.Must(template.ParseGlob("templates/**/*.html"))

    // Configurar rotas
    mux := http.NewServeMux()
    
    // Rotas estáticas
    mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    
    // Rotas de autenticação
    mux.HandleFunc("/login", handlers.LoginHandler(db, tmpl))
    mux.HandleFunc("/logout", handlers.LogoutHandler())
    mux.HandleFunc("/register", handlers.RegisterHandler(db, tmpl))
    
    // Rotas de agendamento
    mux.HandleFunc("/appointments", AuthMiddleware(handlers.AppointmentsHandler(db, tmpl)))
    mux.HandleFunc("/appointments/new", AuthMiddleware(handlers.NewAppointmentHandler(db, tmpl)))
    mux.HandleFunc("/appointments/accept", AuthMiddleware(handlers.AcceptAppointmentHandler(db)))
    mux.HandleFunc("/appointments/reject", AuthMiddleware(handlers.RejectAppointmentHandler(db)))
    
    // WebSocket
    mux.HandleFunc("/ws", AuthMiddleware(handlers.WebSocketHandler(db)))

    // Iniciar servidor
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

### Autenticação (auth.go)
```go
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session")
        if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        next(w, r)
    }
}

func LoginHandler(db *sql.DB, tmpl *template.Template) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            email := r.FormValue("email")
            password := r.FormValue("password")
            
            var user User
            err := db.QueryRow(`
                SELECT id, password_hash 
                FROM users 
                WHERE email = $1`, email).Scan(&user.ID, &user.PasswordHash)
            
            if err == nil && checkPasswordHash(password, user.PasswordHash) {
                session, _ := store.Get(r, "session")
                session.Values["authenticated"] = true
                session.Values["user_id"] = user.ID
                session.Save(r, w)
                http.Redirect(w, r, "/appointments", http.StatusSeeOther)
                return
            }
            
            http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
            return
        }
        
        tmpl.ExecuteTemplate(w, "login.html", nil)
    }
}
```

### WebSocket Simplificado (notifications.go)
```go
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func WebSocketHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID := r.Context().Value("user_id").(int)
        
        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            log.Printf("Erro no WebSocket: %v", err)
            return
        }
        
        client := &Client{
            UserID: userID,
            Conn:   conn,
        }
        
        clients[userID] = client
        defer func() {
            delete(clients, userID)
            conn.Close()
        }()
        
        for {
            _, _, err := conn.ReadMessage()
            if err != nil {
                break
            }
        }
    }
}

func SendNotification(userID int, notification Notification) {
    // Salvar no banco
    _, err := db.Exec(`
        INSERT INTO notifications (user_id, title, message, type)
        VALUES ($1, $2, $3, $4)`,
        userID, notification.Title, notification.Message, notification.Type)
    if err != nil {
        log.Printf("Erro ao salvar notificação: %v", err)
        return
    }
    
    // Enviar via WebSocket
    if client, ok := clients[userID]; ok {
        client.Conn.WriteJSON(notification)
    }
    
    // Enviar email assíncrono
    go sendEmail(userID, notification)
}
```
````



Continuando com a próxima parte do markdown:

````markdown
### Templates HTML

#### Layout Base (layout/base.html)
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Superviso - {{.Title}}</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link href="/static/css/style.css" rel="stylesheet">
    <script src="https://unpkg.com/htmx.org@1.9.0"></script>
</head>
<body>
    <nav class="navbar navbar-expand-lg navbar-dark bg-dark">
        <div class="container">
            <a class="navbar-brand" href="/">Superviso</a>
            {{if .User}}
                <div class="navbar-nav">
                    <a class="nav-link" href="/appointments">Agendamentos</a>
                    <a class="nav-link" href="/logout">Sair</a>
                </div>
            {{end}}
        </div>
    </nav>

    <div class="container mt-4">
        {{template "content" .}}
    </div>

    <div id="notifications" class="position-fixed top-0 end-0 p-3"></div>

    <script src="/static/js/app.js"></script>
</body>
</html>
```

#### Lista de Agendamentos (appointments/list.html)
```html
{{define "content"}}
<div class="row">
    <div class="col-md-12">
        <h2>Meus Agendamentos</h2>
        
        <div class="mb-4">
            <button class="btn btn-primary" hx-get="/appointments/new" hx-target="#main">
                Novo Agendamento
            </button>
        </div>

        <div id="appointments-list" hx-get="/appointments/data" hx-trigger="load, every 30s">
            {{range .Appointments}}
                <div class="card mb-3">
                    <div class="card-body">
                        <h5 class="card-title">
                            {{if .IsSupervisor}}
                                {{.SuperviseeName}}
                            {{else}}
                                {{.SupervisorName}}
                            {{end}}
                        </h5>
                        <p class="card-text">
                            Data: {{.Date.Format "02/01/2006"}}
                            <br>
                            Horário: {{.StartTime.Format "15:04"}} - {{.EndTime.Format "15:04"}}
                        </p>
                        {{if eq .Status "pending"}}
                            {{if .IsSupervisor}}
                                <button class="btn btn-success btn-sm" 
                                        hx-post="/appointments/{{.ID}}/accept"
                                        hx-confirm="Confirmar este agendamento?">
                                    Aceitar
                                </button>
                                <button class="btn btn-danger btn-sm"
                                        hx-post="/appointments/{{.ID}}/reject"
                                        hx-confirm="Rejeitar este agendamento?">
                                    Rejeitar
                                </button>
                            {{else}}
                                <span class="badge bg-warning">Aguardando Confirmação</span>
                            {{end}}
                        {{else}}
                            <span class="badge bg-{{.StatusColor}}">{{.StatusText}}</span>
                        {{end}}
                    </div>
                </div>
            {{else}}
                <div class="alert alert-info">
                    Nenhum agendamento encontrado.
                </div>
            {{end}}
        </div>
    </div>
</div>
{{end}}
```

### JavaScript (app.js)
```javascript
// Configuração do WebSocket
class NotificationManager {
    constructor() {
        this.ws = new WebSocket(`ws://${window.location.host}/ws`)
        this.setupWebSocket()
    }

    setupWebSocket() {
        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data)
            
            switch(data.type) {
                case 'notification':
                    this.showNotification(data)
                    break
                case 'appointment_update':
                    this.handleAppointmentUpdate(data)
                    break
            }
        }

        this.ws.onclose = () => {
            console.log('WebSocket fechado. Reconectando...')
            setTimeout(() => new NotificationManager(), 1000)
        }
    }

    showNotification(data) {
        const toast = document.createElement('div')
        toast.className = 'toast show'
        toast.innerHTML = `
            <div class="toast-header">
                <strong class="me-auto">${data.title}</strong>
                <button type="button" class="btn-close" onclick="this.parentElement.parentElement.remove()"></button>
            </div>
            <div class="toast-body">${data.message}</div>
        `
        document.getElementById('notifications').appendChild(toast)
        setTimeout(() => toast.remove(), 5000)
    }

    handleAppointmentUpdate(data) {
        const appointmentsList = document.getElementById('appointments-list')
        if (appointmentsList) {
            htmx.trigger(appointmentsList, 'refresh')
        }
    }
}

// Inicialização
document.addEventListener('DOMContentLoaded', () => {
    new NotificationManager()
})
```

### CSS (style.css)
```css
/* Notificações */
.toast {
    background: white;
    border-radius: 4px;
    box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.15);
    margin-bottom: 1rem;
    max-width: 350px;
}

/* Cards de Agendamento */
.appointment-card {
    transition: all 0.3s ease;
}

.appointment-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 0.5rem 1rem rgba(0, 0, 0, 0.1);
}

/* Status */
.status-pending { background-color: #ffc107; }
.status-confirmed { background-color: #28a745; }
.status-rejected { background-color: #dc3545; }
.status-cancelled { background-color: #6c757d; }

/* Responsividade */
@media (max-width: 768px) {
    .container { padding: 1rem; }
    .toast { max-width: 100%; margin: 0.5rem; }
}
```
````



Continuando com a próxima parte do markdown:

````markdown
## Fluxos de Usuário

### 1. Fluxo de Agendamento
1. **Supervisionando**:
   - Visualiza horários disponíveis
   - Seleciona horário
   - Confirma agendamento
   - Recebe notificação de pendente
   - Aguarda confirmação

2. **Supervisor**:
   - Recebe notificação de nova solicitação
   - Visualiza detalhes do agendamento
   - Aceita ou rejeita
   - Recebe confirmação da ação

### 2. Sistema de Notificações
1. **Tipos de Mensagem**:
```go
const (
    NotificationTypeNewRequest     = "new_request"
    NotificationTypeConfirmation   = "confirmation"
    NotificationTypeRejection      = "rejection"
    NotificationTypeReminder       = "reminder"
)
```

2. **Estrutura da Mensagem**:
```go
type NotificationMessage struct {
    Type      string    `json:"type"`
    Title     string    `json:"title"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
}
```

### 3. Tratamento de Erros
```go
// Erros comuns
var (
    ErrSlotNotAvailable = errors.New("horário não disponível")
    ErrUnauthorized     = errors.New("não autorizado")
    ErrInvalidInput     = errors.New("dados inválidos")
)

// Handler de erro
func handleError(w http.ResponseWriter, err error) {
    switch err {
    case ErrSlotNotAvailable:
        http.Error(w, err.Error(), http.StatusConflict)
    case ErrUnauthorized:
        http.Error(w, err.Error(), http.StatusUnauthorized)
    case ErrInvalidInput:
        http.Error(w, err.Error(), http.StatusBadRequest)
    default:
        log.Printf("Erro interno: %v", err)
        http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
    }
}
```

## Segurança

### 1. Autenticação
```go
// Middleware de autenticação
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        session, _ := store.Get(r, "session")
        if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Validação de senha
func validatePassword(password string) bool {
    return len(password) >= 8
}

// Hash de senha
func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}
```

### 2. Proteção contra CSRF
```go
func setupCSRF() {
    csrf.Protect([]byte(os.Getenv("CSRF_KEY")),
        csrf.Secure(true),
        csrf.HttpOnly(true),
    )
}
```

### 3. Headers de Segurança
```go
func securityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        next.ServeHTTP(w, r)
    })
}
```

## Testes

### 1. Testes de Unidade
```go
func TestCreateAppointment(t *testing.T) {
    tests := []struct {
        name          string
        input         AppointmentInput
        expectedError error
    }{
        {
            name: "valid appointment",
            input: AppointmentInput{
                SupervisorID: 1,
                Date:        time.Now().Add(24 * time.Hour),
                StartTime:   "14:00",
            },
            expectedError: nil,
        },
        {
            name: "past date",
            input: AppointmentInput{
                SupervisorID: 1,
                Date:        time.Now().Add(-24 * time.Hour),
                StartTime:   "14:00",
            },
            expectedError: ErrInvalidInput,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := createAppointment(tt.input)
            if err != tt.expectedError {
                t.Errorf("got %v, want %v", err, tt.expectedError)
            }
        })
    }
}
```

### 2. Testes de Integração
```go
func TestAppointmentFlow(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    // Criar supervisor
    supervisorID := createTestUser(t, db, true)
    
    // Criar supervisionando
    superviseeID := createTestUser(t, db, false)
    
    // Criar agendamento
    appointmentID := createTestAppointment(t, db, supervisorID, superviseeID)
    
    // Verificar status inicial
    status := getAppointmentStatus(t, db, appointmentID)
    if status != "pending" {
        t.Errorf("got status %s, want pending", status)
    }
    
    // Aceitar agendamento
    acceptAppointment(t, db, appointmentID)
    
    // Verificar status final
    status = getAppointmentStatus(t, db, appointmentID)
    if status != "confirmed" {
        t.Errorf("got status %s, want confirmed", status)
    }
}
```

## Deploy e Manutenção

### 1. Variáveis de Ambiente
```bash
# .env
DATABASE_URL=postgres://user:pass@localhost:5432/superviso
SESSION_KEY=your-secret-key
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your@email.com
SMTP_PASS=your-password
```

### 2. Backup do Banco
```bash
#!/bin/bash
# backup.sh
pg_dump superviso > "backup_$(date +%Y%m%d).sql"
```

### 3. Monitoramento
```go
func setupMonitoring() {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        if err := db.Ping(); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
    })
}
```

## Próximos Passos

1. **Melhorias Futuras**
   - Sistema de feedback após supervisão
   - Relatórios e analytics
   - Integração com calendário
   - App mobile

2. **Otimizações**
   - Cache de consultas frequentes
   - Compressão de assets
   - CDN para arquivos estáticos
   - Otimização de queries

3. **Novas Funcionalidades**
   - Chat integrado
   - Pagamentos online
   - Videoconferência
   - Sistema de avaliação
````


