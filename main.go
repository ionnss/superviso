package main

import (
	"log"
	"net/http"
	"superviso/api/routes"
	"superviso/db"

	"github.com/gorilla/mux"
)

func main() {
	// Conecta ao banco de dados
	conn, err := db.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer conn.Close()

	// Configura o roteador
	r := mux.NewRouter()
	routes.ConfigureRoutes(r, conn)

	// Inicia o servidor
	log.Println("Servidor rodando na porta :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
