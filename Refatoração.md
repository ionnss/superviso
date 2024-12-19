```markdown
# Superviso - Refatoração

## Estrutura do Projeto
```
/superviso
├── main.go              # Configuração e inicialização do servidor
├── handlers/            # Handlers HTTP
│   ├── auth.go         # Autenticação e sessões
│   ├── appointments.go  # Agendamentos
│   └── notifications.go # Notificações e WebSocket
├── templates/          # Templates HTML
│   ├── layout.html    # Template base
│   ├── auth/         # Templates de autenticação
│   └── appointments/ # Templates de agendamentos
├── static/           # Assets estáticos
│   ├── css/         # Estilos
│   ├── js/          # JavaScript
│   └── assets/      # Imagens e outros recursos
└── db/              # Migrations e queries SQL
```

## Banco de Dados

### Tabelas
```sql
-- Usuários
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Perfis de Supervisor
CREATE TABLE supervisor_profiles (
    user_id INT PRIMARY KEY REFERENCES users(id),
    bio TEXT,
    specialties TEXT[]
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
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
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
```

## Handlers Principais

### Autenticação (auth.go)
```go
func LoginHandler(w http.ResponseWriter, r *http.Request)
func LogoutHandler(w http.ResponseWriter, r *http.Request)
func RegisterHandler(w http.ResponseWriter, r *http.Request)
func AuthMiddleware(next http.Handler) http.Handler
```

### Agendamentos (appointments.go)
```go
func ListAppointmentsHandler(w http.ResponseWriter, r *http.Request)
func CreateAppointmentHandler(w http.ResponseWriter, r *http.Request)
func AcceptAppointmentHandler(w http.ResponseWriter, r *http.Request)
func RejectAppointmentHandler(w http.ResponseWriter, r *http.Request)
func CancelAppointmentHandler(w http.ResponseWriter, r *http.Request)
```

### Notificações (notifications.go)
```go
func NotificationsHandler(w http.ResponseWriter, r *http.Request)
func WebSocketHandler(w http.ResponseWriter, r *http.Request)
func SendNotification(userID int, notification Notification)
```

## WebSocket Simplificado
```go
type Client struct {
    UserID int
    Conn   *websocket.Conn
}

var clients = make(map[int]*Client)

func broadcast(message []byte) {
    for _, client := range clients {
        client.Conn.WriteMessage(websocket.TextMessage, message)
    }
}

func sendToUser(userID int, message []byte) {
    if client, ok := clients[userID]; ok {
        client.Conn.WriteMessage(websocket.TextMessage, message)
    }
}
```

## Templates HTML
```html
<!-- layout.html -->
<!DOCTYPE html>
<html>
<head>
    <title>Superviso</title>
    <link href="/static/css/style.css" rel="stylesheet">
    <script src="/static/js/app.js" defer></script>
</head>
<body>
    <nav>{{ template "nav" . }}</nav>
    <main>{{ template "content" . }}</main>
    <div id="notifications"></div>
</body>
</html>

<!-- appointments/list.html -->
<div hx-get="/appointments" hx-trigger="every 30s">
    {{range .Appointments}}
        <div class="appointment-card">
            <h3>{{.Date.Format "02/01/2006"}}</h3>
            <p>{{.StartTime}} - {{.EndTime}}</p>
            {{if eq .Status "pending"}}
                <button hx-post="/appointments/{{.ID}}/accept">Aceitar</button>
                <button hx-post="/appointments/{{.ID}}/reject">Rejeitar</button>
            {{end}}
        </div>
    {{end}}
</div>
```

## JavaScript Simplificado (app.js)
```javascript
// Websocket para notificações
const ws = new WebSocket(`ws://${window.location.host}/ws`)

ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    
    switch(data.type) {
        case 'notification':
            showNotification(data)
            break
        case 'appointment_update':
            htmx.trigger('#appointments', 'refresh')
            break
    }
}

function showNotification(data) {
    const toast = document.createElement('div')
    toast.className = 'toast'
    toast.textContent = data.message
    document.getElementById('notifications').appendChild(toast)
    setTimeout(() => toast.remove(), 5000)
}
```

## CSS Simplificado (style.css)
```css
/* Usando Bootstrap como base */
@import 'bootstrap/dist/css/bootstrap.min.css';

/* Customizações */
.appointment-card {
    border: 1px solid #ddd;
    padding: 1rem;
    margin-bottom: 1rem;
    border-radius: 4px;
}

.toast {
    position: fixed;
    top: 1rem;
    right: 1rem;
    background: #333;
    color: white;
    padding: 1rem;
    border-radius: 4px;
    z-index: 1000;
}
```

## Funcionalidades Mantidas
1. Autenticação de usuários
2. Perfis de supervisor
3. Agendamento de supervisões
4. Aprovação/Rejeição de agendamentos
5. Notificações em tempo real
6. Notificações por email
7. Histórico de agendamentos
8. Visualização de disponibilidade

## Simplificações Principais
1. Menos estados de agendamento
2. WebSocket mais simples
3. Sem camada de models separada
4. Queries SQL diretas
5. Um arquivo JS principal
6. Um arquivo CSS principal
7. Menos dependências
8. Estrutura de diretórios mais plana

## Tecnologias Mantidas
- Go para backend
- HTMX para interatividade
- WebSocket para tempo real
- Bootstrap para UI base
- PostgreSQL para banco de dados

## Próximos Passos
1. Implementar sistema de feedback
2. Adicionar lembretes automáticos
3. Melhorar relatórios
4. Adicionar filtros de busca
5. Implementar sistema de cancelamento
```

Quer que eu detalhe alguma parte específica?
