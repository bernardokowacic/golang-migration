package main

import (
	"log"
	"net/http"

	"github.com/golang-migration/routes"
	"github.com/joho/godotenv"

	_ "github.com/denisenkom/go-mssqldb"
)

func init() {
	log.SetPrefix("LOG: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("init started")

	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {

	// qq, _ := models.GetAllMigrations(0, 15)

	// fmt.Println(qq.Items)

	routes.LoadRoutes()
	http.ListenAndServe(":8000", nil)
}
