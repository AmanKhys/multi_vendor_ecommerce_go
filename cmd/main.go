package main

import (
	"fmt"
	"strconv"
	"time"

	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/envname"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/router"

	env "github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type config struct {
	port int
	env  string
}

// var dbConn = repository.NewDBConfig()
// var DB = db.New(dbConn)

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
		env:  envname.Development,
	}

	mux := router.SetupRouter()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	// remove void orders and payments according to it;
	go OrdersCron()
	log.Printf("Server running on port %d in %s", cfg.port, cfg.env)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
