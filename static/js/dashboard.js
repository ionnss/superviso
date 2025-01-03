document.addEventListener('DOMContentLoaded', function() {
    const roleWarning = document.getElementById('roleWarning');
    if (roleWarning) {
        if (localStorage.getItem('roleWarningDismissed')) {
            roleWarning.style.display = 'none';
        }
    }

    // Only run navbar collapse code if navbar exists
    const navbar = document.getElementById('navbarResponsive');
    const toggler = document.querySelector('.navbar-toggler');
    
    if (navbar && toggler) {
        document.addEventListener('click', function(e) {
            if (navbar.classList.contains('show') && 
                !navbar.contains(e.target) && 
                !toggler.contains(e.target)) {
                bootstrap.Collapse.getInstance(navbar).hide();
            }
        });
    }
});

// Listen for HTMX after-request to maintain state
document.addEventListener('htmx:afterRequest', function(evt) {
    if (evt.detail.pathInfo.requestPath === '/api/profile/toggle-supervisor') {
        const supervisorSettings = document.getElementById('supervisorSettings');
        const supervisorToggle = document.getElementById('supervisorToggle');
        if (supervisorSettings && supervisorToggle) {
            supervisorSettings.style.display = supervisorToggle.checked ? 'block' : 'none';
        }
    }
});

// Função para verificar o papel do usuário via API
function checkUserRole() {
    fetch('/api/profile/check-role')
        .then(response => response.json())
        .then(data => {
            const roleWarning = document.getElementById('roleWarning');
            if (roleWarning) {
                // Verificar idade da conta
                fetch('/api/profile/check-age')
                    .then(response => response.json())
                    .then(ageData => {
                        roleWarning.style.display = 
                            (data.hasRole || ageData.isOldEnough) ? 'none' : 'block';
                    });
            }
        });
}

// Verificar quando a página carrega
document.addEventListener('DOMContentLoaded', checkUserRole);

// Verificar quando o perfil é atualizado (evento HTMX)
document.addEventListener('htmx:afterRequest', function(evt) {
    if (evt.detail.pathInfo.requestPath === '/api/profile/update') {
        checkUserRole();
    }
});
