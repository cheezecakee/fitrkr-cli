package main

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx driver
)

func NewConnection(connString string) *sql.DB {
	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	return db
}
