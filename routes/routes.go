package routes

import (
	"net/http"

	"github.com/golang-migration/controllers"
)

// LoadRoutes ... Carrega rotas disponíveis no sistema
func LoadRoutes() {
	http.HandleFunc("/", controllers.Index)
	http.HandleFunc("/update-production", controllers.UpdateProduction)
	http.HandleFunc("/update-test", controllers.UpdateTest)
	http.HandleFunc("/save-migration", controllers.SaveMigration)
	http.HandleFunc("/delete-migration", controllers.DeleteMigration)
	http.HandleFunc("/show-logs", controllers.ShowLogs)
}
