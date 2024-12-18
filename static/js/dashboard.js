document.addEventListener('DOMContentLoaded', function() {
    const roleWarning = document.getElementById('roleWarning');
    if (roleWarning) {
        if (localStorage.getItem('roleWarningDismissed')) {
            roleWarning.style.display = 'none';
        }
    }
});

function hideRoleWarning() {
    localStorage.setItem('roleWarningDismissed', 'true');
    document.getElementById('roleWarning').style.display = 'none';
} 