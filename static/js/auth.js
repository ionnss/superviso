// Funções relacionadas à autenticação
function handleLoginResponse(event) {
    console.log('Login Response Event:', event.detail);
    const messageContainer = document.getElementById('message-container');
    
    // Se for sucesso (status 200)
    if (event.detail.xhr.status === 200) {
        messageContainer.innerHTML = `
            <div class="alert alert-success">
                <i class="fas fa-check-circle me-2"></i>
                Login realizado com sucesso! Redirecionando...
            </div>
        `;
        setTimeout(() => window.location.href = '/dashboard', 1500);
        return;
    }
    
    // Verifica se é erro de email não verificado
    const triggerHeader = event.detail.xhr.getResponseHeader('HX-Trigger');
    if (triggerHeader) {
        try {
            const trigger = JSON.parse(triggerHeader);
            if (trigger.showVerification) {
                messageContainer.innerHTML = `
                    <div class="alert alert-warning" style="background-color: #fcf0bd; color: #664d03; border-color: #f7dd7e;">
                        <i class="fas fa-exclamation-circle me-2"></i>
                        Por favor, confirme seu email antes de fazer login.<br>
                        <button class="btn btn-link p-0 mt-2" style="color: #664d03; text-decoration-color: #664d03;" onclick="resendVerification('${trigger.showVerification.email}')">
                            <i class="fas fa-envelope me-1"></i>
                            Reenviar email de confirmação
                        </button>
                    </div>
                `;
                return;
            }
        } catch (e) {
            console.error('Erro ao parsear HX-Trigger:', e);
        }
    }
    
    // Se for erro 401 (não autorizado)
    if (event.detail.xhr.status === 401) {
        messageContainer.innerHTML = `
            <div class="alert alert-danger">
                <i class="fas fa-exclamation-circle me-2"></i>
                Email ou senha incorretos
            </div>
        `;
        return;
    }
    
    // Se for erro 500 ou outro erro
    messageContainer.innerHTML = `
        <div class="alert alert-danger">
            <i class="fas fa-exclamation-circle me-2"></i>
            Erro ao fazer login. Por favor, tente novamente.
        </div>
    `;
}

// Debug events - Remover em produção
document.body.addEventListener('htmx:afterRequest', function(event) {
    console.log('HTMX Response:', {
        status: event.detail.xhr.status,
        successful: event.detail.successful,
        headers: event.detail.xhr.getAllResponseHeaders(),
        response: event.detail.xhr.response
    });
}); 