package main

import (
	"log"
	"net/http"

	"github.com/golang-migration/routes"
	"github.com/joho/godotenv"

	_ "github.com/denisenkom/go-mssqldb"
)

func main() {
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatal("Error loading .env file")
	}

	routes.LoadRoutes()
	http.ListenAndServe(":8000", nil)
}
