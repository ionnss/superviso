// Manipulação do modal de agendamento
document.addEventListener('show.bs.modal', function (event) {
    if (event.target.id === 'confirmModal') {
        const button = event.relatedTarget;
        const slotId = button.getAttribute('data-slot-id');
        const slotDate = button.getAttribute('data-slot-date');
        const slotTime = button.getAttribute('data-slot-time');
        
        document.getElementById('modalDate').textContent = slotDate;
        document.getElementById('modalTime').textContent = slotTime;
        document.getElementById('confirmBooking').setAttribute('hx-vals', `{"slot_id": ${slotId}}`);
    }
}); 