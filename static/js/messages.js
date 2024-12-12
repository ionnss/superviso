// Sistema de mensagens
function initializeSystemMessages() {
    const systemMessage = document.getElementById('system-message');
    if (!systemMessage) return;

    // Usa o URLSearchParams global do index.html
    const msg = window.location.search ? new URLSearchParams(window.location.search).get('msg') : null;
    if (!msg) return;

    const messages = {
        'register_success': `
            <div class="alert alert-success alert-dismissible fade show">
                <i class="fas fa-check-circle me-2"></i>
                Cadastro realizado com sucesso! Por favor, faça login.
                <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
            </div>
        `,
        'logout_success': `
            <div class="alert alert-success alert-dismissible fade show">
                <i class="fas fa-check-circle me-2"></i>
                Logout realizado com sucesso!
                <button type="button" class="btn-close" data-bs-dismiss="alert" id="logoutAlert"></button>
                <script>
                    setTimeout(() => {
                        document.querySelector('#logoutAlert').click();
                    }, 5000);
                </script>
            </div>
        `
    };

    if (messages[msg]) {
        systemMessage.innerHTML = messages[msg];
    }
}

// Inicializa as mensagens quando o DOM estiver pronto
document.addEventListener('DOMContentLoaded', initializeSystemMessages); 