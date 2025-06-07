package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"net/http"

	"github.com/amankhys/multi_vendor_ecommerce_go/cmd/crons"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/envname"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/router"

	log "github.com/sirupsen/logrus"
)

type config struct {
	port int
	env  string
}

func main() {
	var cfg config
	portStr := os.Getenv(envname.Port)
	port, err := strconv.Atoi(portStr)
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
	go crons.OrdersCron()
	// clear out vendor_payments accordingly
	go crons.PaymentRoutine()
	log.Printf("Server running on port %d in %s", cfg.port, cfg.env)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
