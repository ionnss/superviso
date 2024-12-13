// Funções relacionadas à verificação de email
function resendVerification(email) {
    const messageContainer = document.getElementById('message-container');
    messageContainer.innerHTML = `
        <div class="alert alert-info d-flex align-items-center">
            <div class="d-flex align-items-center">
                <span class="spinner-border spinner-border-sm me-2"></span>
                <span>Reenviando email de verificação para ${email}...</span>
            </div>
        </div>
    `;
    
    htmx.ajax('POST', '/resend-verification', {
        target: '#message-container',
        swap: 'innerHTML',
        values: { email: email }
    });
} 