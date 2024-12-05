// superviso/db/connection.go
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // Driver PostgreSQL
)

var DB *sql.DB

// Connect inicializa a conexão com o banco de dados e a retorna
func Connect() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	// Tenta conectar ao banco de dados
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Verifica se a conexão está funcionando
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// ExecuteMigrations executa os scripts de migração no banco
func ExecuteMigrations(conn *sql.DB) {
	files := []string{
		"db/migrations/create_users_table.sql",
		// Adicione outros scripts de migração aqui, se necessário
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Erro ao ler o arquivo de migração %s: %v", file, err)
		}

		_, err = conn.Exec(string(content))
		if err != nil {
			log.Fatalf("Erro ao executar o script de migração %s: %v", file, err)
		}

		log.Printf("Migração executada com sucesso: %s", file)
	}
}
