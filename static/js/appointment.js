// Manipulação do modal de agendamento
document.addEventListener('show.bs.modal', function (event) {
    if (event.target.id === 'confirmModal') {
        const button = event.relatedTarget;
        const slotId = button.getAttribute('data-slot-id');
        const slotDate = button.getAttribute('data-slot-date');
        const slotTime = button.getAttribute('data-slot-time');
        
        // Atualizar conteúdo do modal
        document.getElementById('modalDate').textContent = slotDate;
        document.getElementById('modalTime').textContent = slotTime;
        
        // Configurar o botão de confirmação
        const confirmButton = document.getElementById('confirmBooking');
        confirmButton.setAttribute('hx-vals', JSON.stringify({
            slot_id: slotId
        }));
    }
});

// Manipulação das ações de aceitar/rejeitar agendamento
document.addEventListener('click', function(e) {
    if (e.target.classList.contains('accept-btn')) {
        const id = e.target.dataset.id;
        htmx.ajax('POST', `/api/appointments/accept?id=${id}`, {
            target: '#main-content',
            swap: 'innerHTML',
            afterRequest: function() {
                // Ativar a aba de confirmados
                document.querySelector('#confirmed-tab').click();
            }
        });
    } else if (e.target.classList.contains('reject-btn')) {
        const id = e.target.dataset.id;
        htmx.ajax('POST', `/api/appointments/reject?id=${id}`, {
            target: '#main-content',
            swap: 'innerHTML'
        });
    }
});

// Adicionar ao final do arquivo
document.addEventListener('DOMContentLoaded', function() {
    // Ativar as tabs do Bootstrap
    var triggerTabList = [].slice.call(document.querySelectorAll('#appointmentTabs a'))
    triggerTabList.forEach(function (triggerEl) {
        new bootstrap.Tab(triggerEl)
    })
}); 