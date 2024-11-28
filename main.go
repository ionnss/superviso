// superviso/main.go
package main

import (
	"log"

	"superviso/db"
)

func main() {
	// Connect to databse
	db.Connect()

	// Closes connection when exit
	defer db.DB.Close()

	log.Println("Aplicação iniciada...")
}
