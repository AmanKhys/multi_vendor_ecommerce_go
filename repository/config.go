package repository

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"

	env "github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func NewDBConfig() *sql.DB {
	var envM, err = env.Read(".env")
	if err != nil {
		log.Fatal("error loading environment vairables: ", err)
	}
	var dbName = envM["dbname"]
	var dbPort = envM["dbPort"]
	var dbDriver = envM["dbDriver"]
	var host = envM["host"]
	var dbUser = envM["dbUser"]
	var pw = envM["dbpassword"]

	var connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, pw, host, dbPort, dbName)
	db, err := sql.Open(dbDriver, connStr)
	if err != nil {
		log.Fatal("error connecting to database: ", err)
	}

	return db
}
