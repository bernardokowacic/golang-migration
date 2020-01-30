package routes

import (
	"net/http"

	"github.com/golang-migration/controllers"
)

func LoadRoutes() {
	http.HandleFunc("/", controllers.Index)
	http.HandleFunc("/update-production", controllers.UpdateProduction)
	http.HandleFunc("/update-test", controllers.UpdateTest)
	http.HandleFunc("/save-migration", controllers.SaveMigration)
}
