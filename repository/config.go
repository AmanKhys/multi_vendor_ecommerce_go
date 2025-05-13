package repository

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/envname"
	_ "github.com/lib/pq"

	log "github.com/sirupsen/logrus"
)

func NewDBConfig() *sql.DB {
	var dbName = os.Getenv(envname.DbName)
	var dbPort = os.Getenv(envname.DbPort)
	var dbDriver = os.Getenv(envname.DbDriver)
	var host = os.Getenv(envname.DbHost)
	var dbUser = os.Getenv(envname.DbUser)
	var pw = os.Getenv(envname.DbPassword)
	var timezone = os.Getenv(envname.DbTimeZone)

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
