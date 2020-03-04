package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"text/template"

	"github.com/golang-migration/models"
)

var temp = template.Must(template.ParseGlob("templates/*.html"))

func Index(w http.ResponseWriter, r *http.Request) {
	queries, err := models.GetAllMigrations(0, 0)
	if err != nil {
		fmt.Println(err)
	}

	totalPagesFloat := float64(len(queries.Items) / 15)
	totalPages := uint32(math.Ceil(totalPagesFloat))
	if totalPages <= 0 {
		totalPages = 1
	}

	fmt.Println(totalPages)

	totalRegisters := 2

	var currentPage uint32 = 1

	transformJSON, _ := json.Marshal(queries.Items)

	var nextPage uint32 = totalPages
	if currentPage < totalPages {
		nextPage = currentPage + 1
	}

	var previousPage uint32 = 1
	if currentPage > 1 {
		previousPage = currentPage - 1
	}

	data := map[string]interface{}{
		"Total_registers": totalRegisters,
		"Current_page":    currentPage,
		"Next_page":       nextPage,
		"Previous_page":   previousPage,
		"Queries":         queries.Items,
		"Json":            string(transformJSON),
	}

	err2 := temp.ExecuteTemplate(w, "Index", data)
	if err2 != nil {
		fmt.Println(err2)
	}
}

func UpdateProduction(w http.ResponseWriter, r *http.Request) {
	queries, err := models.GetAllMigrations(2, 0)
	if err != nil {
		fmt.Println(err)
	}

	if len(queries.Items) < 1 {
		fmt.Println(err)
		fmt.Fprintf(w, "Todas as migrations já foram rodadas no BD de produção")
		return
	}

	for x := 0; x < len(queries.Items); x++ {
		_, err = models.ExecMigration(queries.Items[x], "producao")

		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Erro ao executar a migration")
			return
		}
	}

	fmt.Fprintf(w, "true")
}

func UpdateTest(w http.ResponseWriter, r *http.Request) {
	var selectedMigrations []string //{ value string }
	err := json.Unmarshal([]byte(r.FormValue("migrationsToRun")), &selectedMigrations)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "erro ao receber queries à serem executadas")
		return
	}
	migrationsToRunIDs := strings.Join(selectedMigrations[:], ",")
	fmt.Println(migrationsToRunIDs) // FALTA AJUSTAR O SELECT PARA BUSCAR SOMENTE OS IDS QUE ESTÃO NESSA VARIÁVEL

	queries, err := models.GetAllMigrations(2, 0)
	if err != nil {
		fmt.Println(err)
	}

	if len(queries.Items) < 1 {
		fmt.Println(err)
		fmt.Fprintf(w, "Todas as migrations já foram rodadas no BD de teste")
		return
	}

	for x := 0; x < len(queries.Items); x++ {
		_, err = models.ExecMigration(queries.Items[x], "teste")

		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Erro ao executar a migration")
			return
		}
	}
	fmt.Fprintf(w, "true")
}

func SaveMigration(w http.ResponseWriter, r *http.Request) {
	_, err := models.InsertMigration(r.FormValue("title"), r.FormValue("query"))

	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Erro ao executar insert")
		return
	}

	fmt.Fprintf(w, "true")
}
