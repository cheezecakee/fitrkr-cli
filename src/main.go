package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := os.Setenv("PGAPPNAME", "fitrkrcli"); err != nil {
		log.Fatalf("could not set app name: %v", err)
	}

	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found: %v", err)
	}

	dbConn := os.Getenv("DB_CONN_STRING")
	if dbConn == "" {
		log.Fatal("DB_CONN_STRING environment variable is required")
	}
	log.Println("dbConn: ", dbConn)

	db := NewConnection(dbConn)
	defer db.Close()

	InitMenu(db)
}
