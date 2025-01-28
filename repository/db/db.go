package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"os"
)

func ConnectDB() (*sql.DB, error) {
	host, found := os.LookupEnv("host")
	if !found {
		log.Fatal("host not found in .env")
	}
	user, found := os.LookupEnv("user")
	if !found {
		log.Fatal("user not found in .env")
	}
	pw, found := os.LookupEnv("pw")
	if !found {
		log.Fatal("password not found in .env")
	}
	port, found := os.LookupEnv("dbport")
	if !found {
		log.Fatal("port not found in .env")
	}
	dbname, found := os.LookupEnv("dbname")
	if !found {
		log.Fatal("db name not found in .env")
	}

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
