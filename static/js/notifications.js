// Check if NotificationManager already exists
if (typeof window.NotificationManager === 'undefined') {
    class NotificationManager {
        constructor() {
            this.ws = null;
            this.reconnectAttempts = 0;
            this.maxReconnectAttempts = 5;
            this.reconnectDelay = 1000;
            this.connectionState = 'disconnected';
            this.notificationSound = new Audio('/static/assets/sounds/notification.mp3');
            this.notificationSound.onerror = () => {
                console.warn('Notification sound file not found');
                this.soundEnabled = false;
            };
            this.soundEnabled = localStorage.getItem('notificationSound') !== 'disabled';
            
            // Add CSS for badge transitions
            this.injectStyles();
            
            this.initializeNotifications();
            this.loadInitialBadgeCount();
            this.connect();
            this.initializeMarkAllAsRead();
            this.initializeSoundToggle();
            this.initializeNotificationClickHandlers();
            
            // Debug info
            setInterval(() => {
                console.log('NotificationManager state:', {
                    connectionState: this.connectionState,
                    reconnectAttempts: this.reconnectAttempts,
                    wsReadyState: this.ws ? this.ws.readyState : 'no websocket',
                    soundEnabled: this.soundEnabled
                });
                this.updateNotificationBadge();
            }, 10000);
        }

        injectStyles() {
            const style = document.createElement('style');
            style.textContent = `
                #notificationCount {
                    transition: opacity 0.2s ease-in-out;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    min-width: 18px;
                    height: 18px;
                    padding: 2px 5px;
                    font-size: 0.75rem;
                    font-weight: bold;
                    z-index: 1000;
                }
                
                .notification-item {
                    transition: background-color 0.2s ease-in-out;
                }
                
                .notification-item.unread {
                    background-color: rgba(52, 152, 219, 0.2);
                }
                
                .notification-item.read {
                    background-color: transparent;
                }
            `;
            document.head.appendChild(style);
        }

        loadInitialBadgeCount() {
            console.log('Loading initial badge count');
            this.updateNotificationBadge();
        }

        updateNotificationBadge() {
            console.log('Updating notification badge');
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
                        const unreadCount = parseInt(count) || 0;
                        console.log('Unread count:', unreadCount);
                        
                        // Update badge text and visibility
                        badge.textContent = unreadCount > 0 ? unreadCount.toString() : '';
                        
                        if (unreadCount > 0) {
                            badge.style.display = 'flex';
                            badge.style.opacity = '1';
                        } else {
                            badge.style.display = 'none';
                            badge.style.opacity = '0';
                        }

                        // Dispatch event for other components
                        document.dispatchEvent(new CustomEvent('notifications-updated', {
                            detail: { count: unreadCount }
                        }));
                    } else {
                        console.warn('Notification badge element not found');
                    }
                })
                .catch(error => {
                    console.error('Error updating badge:', error);
                });
        }

        initializeNotificationClickHandlers() {
            // Handle clicks on individual notifications
            document.addEventListener('click', (e) => {
                const notificationItem = e.target.closest('.notification-item');
                if (notificationItem && !notificationItem.classList.contains('read')) {
                    const notificationId = notificationItem.dataset.id;
                    this.markNotificationAsRead(notificationId);
                }
            });
        }

        markNotificationAsRead(notificationId) {
            fetch(`/api/notifications/${notificationId}/read`, {
                method: 'POST',
            })
            .then(response => {
                if (response.ok) {
                    console.log('Notification marked as read:', notificationId);
                    
                    // Trigger HTMX update
                    document.body.dispatchEvent(new Event('notification-read'));
                    
                    // Update badge
                    this.updateNotificationBadge();
                    
                    // Update notification list
                    const notificationItem = document.querySelector(`[data-id="${notificationId}"]`);
                    if (notificationItem) {
                        notificationItem.classList.remove('unread');
                        notificationItem.classList.add('read');
                    }
                }
            })
            .catch(error => {
                console.error('Error marking notification as read:', error);
            });
        }

        initializeMarkAllAsRead() {
            const markAllBtn = document.getElementById('markAllAsRead');
            if (markAllBtn) {
                markAllBtn.addEventListener('click', () => {
                    console.log('Marking all notifications as read');
                    fetch('/api/notifications/mark-all-read', { method: 'POST' })
                        .then(response => {
                            if (response.ok) {
                                // Trigger HTMX updates
                                document.body.dispatchEvent(new Event('notification-read'));
                                
                                // Update badge
                                this.updateNotificationBadge();
                                
                                // Update list
                                htmx.ajax('GET', '/api/notifications', {
                                    target: '#notificationsList',
                                    swap: 'innerHTML'
                                });

                                // Mark all visible notifications as read
                                document.querySelectorAll('.notification-item.unread').forEach(item => {
                                    item.classList.remove('unread');
                                    item.classList.add('read');
                                });
                            }
                        })
                        .catch(error => console.error('Error marking all notifications as read:', error));
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

        // New method to handle notification permissions
        initializeNotifications() {
            // Only request permission when user interacts with the page
            document.addEventListener('click', () => {
                if (!this.hasRequestedPermission && "Notification" in window) {
                    Notification.requestPermission();
                    this.hasRequestedPermission = true;
                }
            }, { once: true }); // Only run once
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
                    icon: "/static/assets/img/logo.svg"
                });
            }
        }

        resetReconnectAttempts() {
            this.reconnectAttempts = 0;
        }

        connect() {
            try {
                this.connectionState = 'connecting';
                const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
                const wsUrl = `${protocol}//${window.location.host}/ws`;
                console.log('Attempting WebSocket connection to:', wsUrl);
                
                this.ws = new WebSocket(wsUrl);
                
                this.ws.onopen = () => {
                    console.log('WebSocket connected successfully');
                    this.connectionState = 'connected';
                    this.resetReconnectAttempts();
                };

                this.ws.onmessage = (event) => {
                    try {
                        const data = JSON.parse(event.data);
                        console.log('WebSocket message received:', data);
                        
                        switch (data.type) {
                            case 'slot_update':
                                this.handleSlotUpdate(data);
                                break;
                            case 'appointment_update':
                                this.handleAppointmentUpdate(data);
                                break;
                            case 'new_appointment':
                            case 'appointment_accepted':
                            case 'appointment_confirmed':
                                this.updateNotificationBadge();
                                this.addNotificationToList(data);
                                if (this.soundEnabled) {
                                    this.playNotificationSound();
                                }
                                this.showBrowserNotification(data);
                                break;
                            default:
                                console.log('Unknown notification type:', data.type);
                                break;
                        }
                    } catch (error) {
                        console.error('Error processing WebSocket message:', error, 'Raw message:', event.data);
                    }
                };

                this.ws.onclose = (event) => {
                    this.connectionState = 'disconnected';
                    console.log('WebSocket closed. Code:', event.code, 'Reason:', event.reason);
                    if (this.reconnectAttempts < this.maxReconnectAttempts) {
                        console.log(`Attempting to reconnect (${this.reconnectAttempts + 1}/${this.maxReconnectAttempts})...`);
                        this.reconnectAttempts++;
                        setTimeout(() => this.connect(), this.reconnectDelay * this.reconnectAttempts);
                    } else {
                        console.error('Maximum reconnection attempts reached');
                    }
                };

                this.ws.onerror = (error) => {
                    this.connectionState = 'error';
                    console.error('WebSocket error:', error);
                };
            } catch (error) {
                this.connectionState = 'error';
                console.error('Failed to establish WebSocket connection:', error);
                if (this.reconnectAttempts < this.maxReconnectAttempts) {
                    setTimeout(() => this.connect(), this.reconnectDelay);
                }
            }
        }

        handleSlotUpdate(data) {
            console.log('Handling slot update:', data);
            // Dispatch event for AppointmentManager
            window.dispatchEvent(new CustomEvent('slotUpdate', { detail: data }));
        }

        handleAppointmentUpdate(data) {
            console.log('Handling appointment update:', data);
            // Dispatch event for AppointmentManager
            window.dispatchEvent(new CustomEvent('appointmentUpdate', { detail: data }));
        }

        addNotificationToList(notification) {
            const list = document.getElementById('notificationsList');
            if (!list) {
                console.warn('Notifications list element not found');
                return;
            }

            console.log('Adding notification to list:', notification);

            htmx.ajax('GET', '/api/notifications', {
                target: '#notificationsList',
                swap: 'innerHTML',
                values: {},
                headers: {
                    'Content-Type': 'application/json'
                }
            });

            this.cleanOldNotifications();
            this.updateNotificationBadge();
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
            if (this.soundEnabled) {
                console.log('Attempting to play notification sound');
                this.notificationSound.play()
                    .then(() => {
                        console.log('Notification sound played successfully');
                    })
                    .catch(err => {
                        console.error('Error playing notification sound:', err);
                        // Try to reload the audio in case it failed to load initially
                        this.notificationSound.load();
                    });
            } else {
                console.log('Notification sound is disabled');
            }
        }
    }

    // Assign to window object to prevent multiple declarations
    window.NotificationManager = NotificationManager;
} 

// Initialize NotificationManager when the DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.notificationManager = new NotificationManager();
    console.log('NotificationManager initialized');
}); 