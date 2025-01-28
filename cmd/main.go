package main

import (
	"fmt"
	"github.com/amankhys/multi_vendor_ecommerce_mvc/pkg/router"
	"github.com/amankhys/multi_vendor_ecommerce_mvc/repository/db"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	config.LoadConfig()

	DB, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	port, found := os.LookupEnv("port")
	if !found {
		log.Fatal("port variable not found")
	}
	r := router.SetupRouter()
	log.Printf("Server running on port %s", port)
	portStr := fmt.Sprintf(":%s", port)
	log.Fatal(http.ListenAndServe(portStr, r))
}
