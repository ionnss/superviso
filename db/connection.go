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

// Package db gerencia a conexão e operações com o banco de dados PostgreSQL.
//
// Fornece:
//   - Conexão com o banco de dados
//   - Execução de migrações
//   - Gerenciamento de transações

// Connect inicializa e retorna uma conexão com o banco de dados.
//
// Utiliza variáveis de ambiente para configuração:
//   - DB_HOST: host do banco
//   - DB_PORT: porta
//   - DB_USER: usuário
//   - DB_PASSWORD: senha
//   - DB_NAME: nome do banco
//
// Retorna:
//   - *sql.DB: conexão com o banco
//   - error: erro se a conexão falhar
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
		"db/migrations/001_create_users_table.sql",
		"db/migrations/002_create_supervisor_profiles_table.sql",
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
