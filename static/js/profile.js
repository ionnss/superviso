document.addEventListener('DOMContentLoaded', function() {
    const supervisorToggle = document.getElementById('supervisorToggle');
    const supervisorSettings = document.getElementById('supervisorSettings');
    
    if (supervisorToggle && supervisorSettings) {
        // Log initial state
        console.log('Initial state:', {
            checked: supervisorToggle.checked,
            display: supervisorSettings.style.display
        });
        
        supervisorSettings.style.display = supervisorToggle.checked ? 'block' : 'none';
        
        supervisorToggle.addEventListener('change', function() {
            // Log toggle change
            console.log('Toggle changed:', {
                checked: this.checked,
                display: supervisorSettings.style.display
            });
            
            supervisorSettings.style.display = this.checked ? 'block' : 'none';
            
            htmx.ajax('POST', '/api/profile/toggle-supervisor', {
                values: { is_supervisor: this.checked },
                headers: {
                    'Content-Type': 'application/json'
                }
            });
        });
    }
}); 