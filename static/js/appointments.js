if (typeof window.AppointmentManager === 'undefined') {
    class AppointmentManager {
        constructor() {
            this.initializeEventListeners();
            this.initializeAutoRefresh();
        }

        initializeEventListeners() {
            // Listen for slot updates
            window.addEventListener('slotUpdate', (event) => {
                this.handleSlotUpdate(event.detail);
            });

            // Listen for appointment updates
            window.addEventListener('appointmentUpdate', (event) => {
                this.handleAppointmentUpdate(event.detail);
            });
        }

        initializeAutoRefresh() {
            // Atualizar lista a cada 5 minutos
            setInterval(() => {
                htmx.trigger('#appointmentsList', 'refresh');
            }, 5 * 60 * 1000);
        }

        handleSlotUpdate(data) {
            console.log('AppointmentManager handling slot update:', data);
            // Atualizar o status do slot na interface
            const slotElement = document.querySelector(`[data-slot-id="${data.slot_id}"]`);
            if (slotElement) {
                slotElement.dataset.status = data.status;
                this.updateSlotAppearance(slotElement, data.status);
            }
        }

        handleAppointmentUpdate(data) {
            console.log('AppointmentManager handling appointment update:', data);
            const appointmentsList = document.getElementById('appointmentsList');
            if (appointmentsList) {
                htmx.trigger(appointmentsList, 'refresh');
            }
        }

        updateSlotAppearance(slotElement, status) {
            // Remove existing status classes
            slotElement.classList.remove('available', 'pending', 'booked');
            // Add new status class
            slotElement.classList.add(status);
            
            // Update button state
            const bookButton = slotElement.querySelector('button');
            if (bookButton) {
                bookButton.disabled = status !== 'available';
                bookButton.textContent = this.getButtonText(status);
            }
        }

        getButtonText(status) {
            switch (status) {
                case 'available':
                    return 'Agendar';
                case 'pending':
                    return 'Pendente';
                case 'booked':
                    return 'Reservado';
                default:
                    return 'Indisponível';
            }
        }
    }
    window.AppointmentManager = AppointmentManager;
}

// Initialize when DOM is loaded and NotificationManager is ready
function initializeAppointmentManager() {
    if (window.notificationManager) {
        console.log('Initializing AppointmentManager');
        window.appointmentManager = new AppointmentManager();
    } else {
        console.log('Waiting for NotificationManager...');
        setTimeout(initializeAppointmentManager, 100);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    initializeAppointmentManager();
}); 