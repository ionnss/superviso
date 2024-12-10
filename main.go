// superviso/main.go
package main

import (
	"log"
	"net/http"
	"superviso/api/routes"
	"superviso/db"

	"github.com/gorilla/mux"
)

// Superviso é uma plataforma de conexão entre supervisores e supervisionados em Psicologia.
//
// Principais características:
//   - Sistema de autenticação seguro
//   - Gerenciamento de perfis de supervisores
//   - Agendamento de supervisões
//   - Backup automático de dados
//   - Interface responsiva com HTMX

func main() {
	// Conecta ao banco de dados
	conn, err := db.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer conn.Close()

	// Executa as migrações
	db.ExecuteMigrations(conn)

	// Configura o roteador
	r := mux.NewRouter()
	routes.ConfigureRoutes(r, conn)

	// Inicia o servidor
	log.Println("Servidor rodando na porta :8080 em http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", r))
}
