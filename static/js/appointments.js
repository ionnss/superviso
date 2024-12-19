class AppointmentManager {
    constructor() {
        this.initializeWebSocket();
        this.initializeAutoRefresh();
    }

    initializeWebSocket() {
        window.notificationManager.ws.addEventListener('message', (event) => {
            const data = JSON.parse(event.data);
            
            switch (data.type) {
                case 'slot_update':
                    this.handleSlotUpdate(data);
                    break;
                case 'appointment_update':
                    this.handleAppointmentUpdate(data);
                    break;
            }
        });
    }

    initializeAutoRefresh() {
        // Atualizar lista a cada 5 minutos
        setInterval(() => {
            htmx.trigger('#appointmentsList', 'refresh');
        }, 5 * 60 * 1000);
    }

    handleSlotUpdate(data) {
        // Atualizar o status do slot na interface
        const slotElement = document.querySelector(`[data-slot-id="${data.slot_id}"]`);
        if (slotElement) {
            slotElement.dataset.status = data.status;
            this.updateSlotAppearance(slotElement, data.status);
        }
    }

    handleAppointmentUpdate(data) {
        // Atualizar a lista de agendamentos
        if (data.status === 'confirmed' || data.status === 'rejected') {
            htmx.trigger('#appointmentsList', 'refresh');
            
            // Limpar toasts antigos
            const toastContainer = document.querySelector('.toast-container');
            Array.from(toastContainer.children).forEach(toast => {
                const bsToast = bootstrap.Toast.getInstance(toast);
                if (bsToast) {
                    bsToast.dispose();
                }
                toast.remove();
            });
            
            // Mostrar toast de feedback
            const toast = document.createElement('div');
            toast.className = `toast align-items-center text-white bg-${data.status === 'confirmed' ? 'success' : 'danger'} border-0`;
            toast.innerHTML = `
                <div class="d-flex">
                    <div class="toast-body">
                        ${data.message}
                    </div>
                    <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
                </div>
            `;
            document.querySelector('.toast-container').appendChild(toast);
            new bootstrap.Toast(toast).show();
        }
    }

    updateSlotAppearance(element, status) {
        // Atualizar classes CSS e texto baseado no status
        element.className = `slot-item ${status}`;
        let statusText = '';
        let buttonClass = '';

        switch (status) {
            case 'available':
                statusText = 'Disponível';
                buttonClass = 'btn-success';
                element.disabled = false;
                break;
            case 'pending':
                statusText = 'Pendente';
                buttonClass = 'btn-warning';
                element.disabled = true;
                break;
            case 'booked':
                statusText = 'Reservado';
                buttonClass = 'btn-secondary';
                element.disabled = true;
                break;
        }

        element.textContent = statusText;
        element.className = `btn ${buttonClass} slot-item ${status}`;
    }
}

// Inicializar quando o DOM estiver pronto
document.addEventListener('DOMContentLoaded', () => {
    window.appointmentManager = new AppointmentManager();
}); 