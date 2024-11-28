// superviso/main.go
package main

import (
	"log"

	"github.com/ionnss/superviso/db"
)

func main() {
	// Connect to databse
	db.Connect()

	// Closes connection when exit
	defer db.DB.Close()

	log.Println("Aplicação iniciada...")
}
