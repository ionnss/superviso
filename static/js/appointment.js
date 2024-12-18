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

// Adicionar listener para o botão de confirmação
document.addEventListener('click', function(event) {
    if (event.target.id === 'confirmBooking') {
        const slotId = JSON.parse(event.target.getAttribute('hx-vals')).slot_id;
        if (!slotId) {
            console.error('Slot ID não encontrado');
            event.preventDefault();
            return;
        }
    }
}); 