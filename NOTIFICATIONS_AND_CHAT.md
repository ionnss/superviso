# Notificações e Chat - Planejamento de Implementação

## 1. Sistema de Notificações por Email

### Eventos que Disparam Emails
- Novo agendamento (para supervisor)
- Confirmação de agendamento (para supervisionando)
- Lembrete 24h antes da supervisão (ambos)
- Cancelamento de sessão (ambos)
- Alteração de horário (ambos)

### Implementação
```go
// email/notifications.go
type SupervisionNotification struct {
    Type          string    // "new_booking", "confirmation", "reminder", etc
    SupervisorID  int
    SuperviseeID  int
    SessionDate   time.Time
    SessionTime   string
}
```

## 2. Dashboard do Supervisionando

### Visualização de Agendamentos
- Lista de supervisões agendadas
- Histórico de supervisões realizadas
- Status de cada sessão
- Opção de cancelamento/remarcação

### Implementação
- Nova view: `view/supervisee/appointments.html`
- Novo endpoint: `GET /api/supervisee/appointments`

## 3. Sistema de Chat

### Websocket
- Conexão em tempo real
- Histórico de mensagens
- Indicador de online/offline
- Notificações de novas mensagens

### Estrutura do Banco
```sql
CREATE TABLE chat_messages (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    supervisee_id INT REFERENCES users(id),
    message TEXT,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    read_at TIMESTAMP
);

CREATE TABLE chat_sessions (
    id SERIAL PRIMARY KEY,
    supervisor_id INT REFERENCES users(id),
    supervisee_id INT REFERENCES users(id),
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP
);
```

### Implementação Backend
```go
// chat/hub.go
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
}

// chat/client.go
type Client struct {
    hub  *Hub
    conn *websocket.Conn
    send chan []byte
}
```

### Implementação Frontend
- Interface de chat em tempo real
- Histórico de conversas
- Indicadores de status
- Notificações desktop

## 4. Ordem de Implementação

1. Sistema de Notificações por Email
   - Configuração do servidor SMTP
   - Templates de email
   - Sistema de fila de emails

2. Dashboard do Supervisionando
   - Interface de agendamentos
   - Sistema de cancelamento
   - Histórico de sessões

3. Sistema de Chat
   - Configuração do Websocket
   - Backend do chat
   - Interface do usuário
   - Sistema de notificações

## 5. Considerações de Segurança

- Autenticação no Websocket
- Criptografia das mensagens
- Validação de permissões
- Rate limiting
- Proteção contra XSS
- Sanitização de mensagens

## 6. Testes Necessários

- Testes de integração do email
- Testes do Websocket
- Testes de carga do chat
- Testes de segurança
- Testes de interface 