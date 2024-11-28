package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // Driver PostgreSQL
)

var DB *sql.DB

// Connect inicializa a conexão com o banco de dados
func Connect() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	// Testa a conexão
	if err = DB.Ping(); err != nil {
		log.Fatalf("Erro ao verificar conexão com o banco de dados: %v", err)
	}

	log.Println("Conexão com o banco de dados estabelecida com sucesso!")
}

// ExecuteMigrations executa os scripts de migração no banco
func ExecuteMigrations() {
	files := []string{
		"db/migrations/create_users_table.sql",
		// Adicione outros scripts de migração aqui, se necessário
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Erro ao ler o arquivo de migração %s: %v", file, err)
		}

		_, err = DB.Exec(string(content))
		if err != nil {
			log.Fatalf("Erro ao executar o script de migração %s: %v", file, err)
		}

		log.Printf("Migração executada com sucesso: %s", file)
	}
}
