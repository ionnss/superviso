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



Vou adicionar uma seção mais detalhada ao arquivo NOTIFICATIONS_AND_CHAT.md, focando em aspectos importantes da implementação:

````markdown
## 7. Detalhamento da Implementação

### 7.1 Sistema de Filas para Notificações
```go
// notifications/queue.go
type NotificationQueue struct {
    queue    chan Notification
    workers  int
    wg       sync.WaitGroup
    shutdown chan struct{}
}

type Notification struct {
    Priority    int       // 1: Alta, 2: Normal, 3: Baixa
    RetryCount  int       // Número de tentativas
    MaxRetries  int       // Máximo de tentativas
    Payload     interface{}
    CreatedAt   time.Time
}

// Worker para processar notificações
func (nq *NotificationQueue) processNotifications() {
    for notification := range nq.queue {
        // Processar com retry exponencial
        // Logging de falhas
        // Métricas de sucesso/falha
    }
}
```

### 7.2 Integração com Calendário
```go
// calendar/integration.go
type CalendarEvent struct {
    Title       string
    Description string
    StartTime   time.Time
    EndTime     time.Time
    Attendees   []string
    Location    string    // Link da sessão virtual
    RemindAt    []time.Duration // [24h, 1h, 15min]
}

// Gerar ICS para diferentes plataformas
func (e *CalendarEvent) GenerateICS() string
func (e *CalendarEvent) GenerateGoogleCalendarLink() string
func (e *CalendarEvent) GenerateOutlookLink() string
```

### 7.3 Sistema de Chat Avançado

#### Recursos do Chat
1. **Salas de Chat Privadas**
```go
type ChatRoom struct {
    ID            string
    SupervisorID  int
    SuperviseeID  int
    Status        string    // active, archived
    CreatedAt     time.Time
    ExpiresAt     time.Time // Auto-arquivamento
    Settings      ChatSettings
}

type ChatSettings struct {
    AllowAttachments bool
    MaxFileSize      int64
    RetentionDays    int
    Notifications    bool
}
```

2. **Sistema de Arquivos**
```go
type FileHandler struct {
    AllowedTypes []string
    MaxSize      int64
    StoragePath  string
}

func (fh *FileHandler) ProcessUpload(file multipart.File) (string, error) {
    // Validar tipo
    // Sanitizar nome
    // Salvar com nome único
    // Gerar URL segura
}
```

3. **Mensagens Estruturadas**
```go
type ChatMessage struct {
    ID        string
    Type      MessageType // text, file, system
    Content   string
    Metadata  MessageMetadata
    Reactions []Reaction
}

type MessageMetadata struct {
    FileType    string
    FileSize    int64
    Duration    int       // Para áudios
    Dimensions  string    // Para imagens
    Preview     string    // URL de preview
}
```

### 7.4 Segurança e Monitoramento

#### Sistema de Logs
```go
type AuditLog struct {
    UserID    int
    Action    string
    Resource  string
    IP        string
    UserAgent string
    Timestamp time.Time
    Success   bool
    Details   map[string]interface{}
}
```

#### Rate Limiting
```go
type RateLimiter struct {
    Limit     int
    Window    time.Duration
    Store     map[string]*TokenBucket
    mu        sync.RWMutex
}

type TokenBucket struct {
    Tokens    float64
    LastCheck time.Time
    Rate      float64
}
```

### 7.5 Métricas e Monitoramento
```go
type ChatMetrics struct {
    ActiveSessions    prometheus.Gauge
    MessagesSent      prometheus.Counter
    ErrorCount        prometheus.Counter
    ResponseTime      prometheus.Histogram
    OnlineUsers       prometheus.Gauge
}

type NotificationMetrics struct {
    QueueSize         prometheus.Gauge
    ProcessingTime    prometheus.Histogram
    FailureRate       prometheus.Counter
    RetryCount        prometheus.Counter
}
```

## 8. Considerações de Escalabilidade

### 8.1 Sharding de Chat
- Distribuição de salas por servidor
- Balanceamento de carga
- Persistência distribuída

### 8.2 Cache
```go
type CacheConfig struct {
    Driver          string    // redis, memcached
    TTL             Duration
    MaxSize         int64
    EvictionPolicy  string
}
```

### 8.3 Backup e Recuperação
- Backup incremental de mensagens
- Exportação de histórico
- Política de retenção
- Recuperação pontual

## 9. Monitoramento e Alertas

### 9.1 Healthchecks
- Latência do WebSocket
- Taxa de erros
- Uso de recursos
- Tempo de resposta

### 9.2 Alertas
- Falhas de entrega
- Picos de uso
- Erros críticos
- Tentativas de violação
````

Este detalhamento adiciona aspectos importantes sobre escalabilidade, monitoramento e segurança. Quer que eu aprofunde em algum aspecto específico?
