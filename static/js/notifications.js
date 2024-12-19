class NotificationManager {
    constructor() {
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // 1 segundo
        this.notificationSound = new Audio('/static/assets/sounds/notification.mp3');
        this.soundEnabled = localStorage.getItem('notificationSound') !== 'disabled';
        this.requestNotificationPermission();
        this.connect();
        this.initializeMarkAllAsRead();
        this.initializeSoundToggle();
    }

    requestNotificationPermission() {
        if ("Notification" in window) {
            Notification.requestPermission();
        }
    }

    showBrowserNotification(notification) {
        if (Notification.permission === "granted" && !document.hasFocus()) {
            new Notification("Superviso", {
                body: notification.message,
                icon: "/static/assets/img/logo.png"
            });
        }
    }

    resetReconnectAttempts() {
        this.reconnectAttempts = 0;
    }

    connect() {
        this.ws = new WebSocket(`ws://${window.location.host}/ws`);
        
        this.ws.onopen = () => {
            console.log('WebSocket conectado');
            this.resetReconnectAttempts();
        };

        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            
            switch (data.type) {
                case 'slot_update':
                    this.handleSlotUpdate(data);
                    break;
                case 'appointment_update':
                    this.handleAppointmentUpdate(data);
                    break;
                default:
                    // Atualizar contador e lista de notificações
                    this.updateNotificationBadge();
                    this.addNotificationToList(data);
                    break;
            }
        };

        this.ws.onclose = () => {
            if (this.reconnectAttempts < this.maxReconnectAttempts) {
                console.log(`WebSocket desconectado. Tentativa ${this.reconnectAttempts + 1} de ${this.maxReconnectAttempts}...`);
                this.reconnectAttempts++;
                setTimeout(() => this.connect(), this.reconnectDelay * this.reconnectAttempts);
            } else {
                console.log('Número máximo de tentativas de reconexão atingido.');
            }
        };

        this.ws.onerror = (error) => {
            console.error('Erro no WebSocket:', error);
        };
    }

    updateNotificationBadge() {
        fetch('/api/notifications/unread-count')
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.text();
            })
            .then(count => {
                const badge = document.getElementById('notificationCount');
                if (badge) {
                    badge.textContent = count > 0 ? count : '';
                    badge.style.display = count > 0 ? 'inline' : 'none';
                }
            })
            .catch(error => {
                console.error('Erro ao atualizar badge:', error);
            });
    }

    addNotificationToList(notification) {
        const list = document.getElementById('notificationsList');
        if (list) {
            // Verificar se já existe uma notificação similar
            const similarNotification = this.findSimilarNotification(notification);
            if (similarNotification) {
                // Atualizar contador na notificação existente
                const counter = similarNotification.querySelector('.notification-counter');
                if (counter) {
                    const count = parseInt(counter.dataset.count || '1') + 1;
                    counter.textContent = `(${count})`;
                    counter.dataset.count = count;
                    return;
                }
            }

            htmx.ajax('GET', '/api/notifications', {target: '#notificationsList'});
            this.cleanOldNotifications();
            this.playNotificationSound();
            this.showBrowserNotification(notification);
        }
    }

    findSimilarNotification(notification) {
        const notifications = document.querySelectorAll('.notification-item');
        const timeThreshold = 5 * 60 * 1000; // 5 minutos

        for (const item of notifications) {
            if (item.dataset.type === notification.type &&
                item.dataset.message === notification.message &&
                (new Date() - new Date(item.dataset.timestamp)) < timeThreshold) {
                return item;
            }
        }
        return null;
    }

    cleanOldNotifications() {
        const oneDayAgo = new Date();
        oneDayAgo.setDate(oneDayAgo.getDate() - 1);
        
        const notifications = document.querySelectorAll('.notification-item');
        notifications.forEach(notification => {
            const timestamp = new Date(notification.dataset.timestamp);
            if (timestamp < oneDayAgo) {
                notification.remove();
            }
        });
    }

    playNotificationSound() {
        // Tocar som apenas se a página não estiver em foco
        if (!document.hasFocus()) {
            this.notificationSound.play().catch(err => {
                console.log('Erro ao tocar som:', err);
            });
        }
    }

    initializeMarkAllAsRead() {
        const markAllBtn = document.getElementById('markAllAsRead');
        if (markAllBtn) {
            markAllBtn.addEventListener('click', () => {
                fetch('/api/notifications/mark-all-read', { method: 'POST' })
                    .then(response => {
                        if (response.ok) {
                            this.updateNotificationBadge();
                            htmx.ajax('GET', '/api/notifications', {target: '#notificationsList'});
                        }
                    })
                    .catch(error => console.error('Erro ao marcar notificações:', error));
            });
        }
    }

    initializeSoundToggle() {
        const toggleBtn = document.getElementById('toggleNotificationSound');
        if (toggleBtn) {
            toggleBtn.addEventListener('click', () => {
                this.soundEnabled = !this.soundEnabled;
                localStorage.setItem('notificationSound', this.soundEnabled ? 'enabled' : 'disabled');
                toggleBtn.innerHTML = this.soundEnabled ? 
                    '<i class="fas fa-volume-up"></i>' : 
                    '<i class="fas fa-volume-mute"></i>';
            });
        }
    }
}

// Inicializar quando o DOM estiver pronto
document.addEventListener('DOMContentLoaded', () => {
    window.notificationManager = new NotificationManager();
}); 