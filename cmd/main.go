package main

import (
	// "fmt"
	"fmt"
	"strconv"
	"time"

	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/router"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository"
	"github.com/amankhys/multi_vendor_ecommerce_go/repository/db"

	env "github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type config struct {
	port int
	env  string
}

var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)

func main() {
	var cfg config
	envM, err := env.Read(".env")
	if err != nil {
		log.Fatal("couldn't load the env", err)
	}
	cfg.port, err = strconv.Atoi(envM["port"])
	if err != nil {
		log.Fatal("invalid port number")
	}
	mux := router.SetupRouter()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Printf("Server running on port %s", cfg.port)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
