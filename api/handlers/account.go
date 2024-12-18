package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"superviso/api/auth"
)

func CheckAccountAge(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(auth.UserIDKey).(int)

		var createdAt time.Time
		err := db.QueryRow(`
			SELECT created_at 
			FROM users 
			WHERE id = $1`, userID).Scan(&createdAt)

		if err != nil {
			http.Error(w, "Erro ao verificar conta", http.StatusInternalServerError)
			return
		}

		accountAge := time.Since(createdAt)
		isOldEnough := accountAge.Hours() >= 48 // 2 dias

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if !isOldEnough {
			// Retorna o mesmo HTML do aviso se a conta não for velha o suficiente
			w.Write([]byte(`
				<div id="roleWarning" class="col-12 mb-4">
					<div class="card border-primary">
						<div class="card-body">
							<div class="d-flex align-items-center">
								<i class="fas fa-info-circle text-primary fa-2x me-3"></i>
								<div>
									<h5 class="card-title mb-2">Defina seu papel no sistema!</h5>
									<p class="card-text mb-2">
										A sua função padrão é <strong>supervisionado</strong>. Caso queira ser supervisor, você deve ativar o modo <strong>supervisor</strong> no seu perfil.
									</p>
									<ul class="mb-3">
										<li><strong>Supervisionado:</strong> Profissional que busca supervisão</li>
										<li><strong>Supervisor:</strong> Profissional que oferece supervisão</li>
									</ul>
									<p class="card-text mb-2">
										Caso queira ser supervisionado, ignore esta mensagem.
									</p>
									<a href="#perfil" 
									   class="btn btn-primary"
									   hx-get="/profile" 
									   hx-target="#main-content">
										<i class="fas fa-user-cog me-2"></i>Configurar Perfil
									</a>
								</div>
							</div>
						</div>
					</div>
				</div>
			`))
		} else {
			// Se a conta for velha o suficiente, não retorna nada
			w.Write([]byte(""))
		}
	}
}
