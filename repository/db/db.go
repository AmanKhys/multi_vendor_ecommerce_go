package db

import (
	"database/sql"
	"fmt"
	env "github.com/joho/godotenv"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	// "os"
)

func ConnectDB() (*sql.DB, error) {
	envM, err := env.Read(".env")
	if err != nil {
		log.Fatal("Couldn't load .env file", err)
	}
	user := envM["user"]
	pw := envM["password"]
	host := envM["dbhost"]
	port := envM["dbport"]
	dbname := envM["dbname"]

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pw, host, port, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging Database: %w", err)
	}
	return db, nil
}
