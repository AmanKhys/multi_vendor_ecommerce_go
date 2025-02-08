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

type EnvType string

const (
	Production  EnvType = "production"
	Development EnvType = "development"
	Testing     EnvType = "testing"
)

type config struct {
	port int
	env  EnvType
}

var dbConn = repository.NewDBConfig()
var DB = db.New(dbConn)

func main() {
	var cfg config
	envM, err := env.Read(".env")
	if err != nil {
		log.Fatal("couldn't load the env", err)
	}
	port, err := strconv.Atoi(envM["port"])
	if err != nil {
		log.Fatal("invalid port number")
	}

	cfg = config{
		port: port,
		env:  Production,
	}

	mux := router.SetupRouter()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	log.Printf("Server running on port %d in %s", cfg.port, cfg.env)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
