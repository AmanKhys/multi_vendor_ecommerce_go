package main

import (
	// "fmt"
	"github.com/amankhys/multi_vendor_ecommerce_go/pkg/router"
	// "github.com/amankhys/multi_vendor_ecommerce_go/repository/db"
	env "github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"net/http"
	// "os"
)

func main() {
	// config.LoadConfig()
	//
	// DB, err := db.ConnectDB()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer DB.Close()

	envM, err := env.Read(".env")
	if err != nil {
		log.Fatal("couldn't load the env", err)
	}
	r := router.SetupRouter()
	port := envM["port"]
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
