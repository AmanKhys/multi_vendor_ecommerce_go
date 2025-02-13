package repository

import (
	"database/sql"
	"fmt"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/envname"
	_ "github.com/lib/pq"

	env "github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func NewDBConfig() *sql.DB {
	var envM, err = env.Read(".env")
	if err != nil {
		log.Fatal("error loading environment vairables: ", err)
	}
	var dbName = envM[envname.DbName]
	var dbPort = envM[envname.DbPort]
	var dbDriver = envM[envname.DbDriver]
	var host = envM[envname.DbHost]
	var dbUser = envM[envname.DbUser]
	var pw = envM[envname.DbPassword]
	var timezone = envM[envname.DbTimeZone]

	var connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&TimeZone=%s", dbUser, pw, host, dbPort, dbName, timezone)
	db, err := sql.Open(dbDriver, connStr)
	if err != nil {
		log.Fatal("error connecting to database: ", err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("error pinging db: ", err)
	}
	log.Info("successful connection to  database;")

	return db
}
