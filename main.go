// Package main é o ponto de entrada da plataforma Superviso.
//
// Superviso é uma plataforma web que conecta psicólogos a supervisores clínicos.
//
// A plataforma oferece:
//   - Sistema de cadastro e autenticação segura
//   - Perfis especializados para supervisores
//   - Busca de supervisores por abordagem e valor
//   - Agendamento de supervisões online
//   - Backup automático de dados
//   - Interface responsiva com HTMX
//
// Principais funcionalidades:
//   - Registro e autenticação de usuários
//   - Gerenciamento de perfis de supervisor
//   - Sistema de busca e filtros
//   - Agendamento de supervisões
//   - Backup automático
package main

import (
	"log"
	"net/http"
	"superviso/api/routes"
	"superviso/db"

	"github.com/gorilla/mux"
)

// main é o ponto de entrada da aplicação.
//
// Inicializa o servidor HTTP, configura o banco de dados e define as rotas.
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
