package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/golang-migration/models"
)

var temp = template.Must(template.ParseGlob("templates/*.html"))

// Index ... Carrega homepage
func Index(w http.ResponseWriter, r *http.Request) {
	var registersPerPage uint32 = 15
	queries, err := models.GetAllMigrations(0, 0, registersPerPage)
	if err != nil {
		fmt.Println(err)
	}
	transformJSON, _ := json.Marshal(queries)

	var currentPage uint32 = 0
	var nextPage uint32 = currentPage + registersPerPage
	var previousPage uint32 = currentPage - registersPerPage
	if previousPage < 0 {
		previousPage = 0
	}

	columnsList, err := models.ShowAllColumns()
	if err != nil {
		fmt.Println(err)
	}
	jsonColumnsList, _ := json.Marshal(columnsList)

	data := map[string]interface{}{
		"Current_page":  currentPage,
		"Next_page":     nextPage,
		"Previous_page": previousPage,
		"Queries":       queries,
		"Columns_list":  string(jsonColumnsList),
		"Json":          string(transformJSON),
	}

	err2 := temp.ExecuteTemplate(w, "Index", data)
	if err2 != nil {
		fmt.Println(err2)
	}
}

// UpdateProduction ... Executa migration no BD de producao
func UpdateProduction(w http.ResponseWriter, r *http.Request) {
	var selectedMigrations []string
	err := json.Unmarshal([]byte(r.FormValue("migrationsToRun")), &selectedMigrations)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "erro ao receber queries à serem executadas")
		return
	}
	if len(selectedMigrations) < 1 {
		fmt.Println("Nenhuma migration selecionada")
		fmt.Fprintf(w, "Nenhuma migration selecionada")
		return
	}
	migrationsToRunIDs := strings.Join(selectedMigrations[:], ",")

	queries, err := models.GetMigrationsByID(2, migrationsToRunIDs)
	if err != nil {
		fmt.Println(err)
	}

	if len(queries) < 1 {
		fmt.Println(err)
		fmt.Fprintf(w, "Todas as migrations já foram rodadas no BD de produção")
		return
	}

	for x := 0; x < len(queries); x++ {
		_, err = models.ExecMigration(queries[x], "producao")

		if err != nil {
			_, errLog := models.CreateMigrationLog(queries[x].Codigo, "produção - "+err.Error())
			if errLog != nil {
				fmt.Println(err)
				fmt.Fprintf(w, "Erro ao criar log de erro da migration")
				return
			}

			fmt.Println(err)
			fmt.Fprintf(w, "Erro ao executar a migration")
			return
		}
	}

	fmt.Fprintf(w, "true")
}

// UpdateTest ... Executa migration no BD de teste
func UpdateTest(w http.ResponseWriter, r *http.Request) {
	var selectedMigrations []string
	err := json.Unmarshal([]byte(r.FormValue("migrationsToRun")), &selectedMigrations)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "erro ao receber queries à serem executadas")
		return
	}
	if len(selectedMigrations) < 1 {
		fmt.Println("Nenhuma migration selecionada")
		fmt.Fprintf(w, "Nenhuma migration selecionada")
		return
	}
	migrationsToRunIDs := strings.Join(selectedMigrations[:], ",")

	queries, err := models.GetMigrationsByID(1, migrationsToRunIDs)
	if err != nil {
		fmt.Println(err)
	}

	if len(queries) < 1 {
		fmt.Println(err)
		fmt.Fprintf(w, "Todas as migrations já foram rodadas no BD de teste")
		return
	}

	for x := 0; x < len(queries); x++ {
		_, err = models.ExecMigration(queries[x], "teste")
		if err != nil {
			_, errLog := models.CreateMigrationLog(queries[x].Codigo, "teste - "+err.Error())
			if errLog != nil {
				fmt.Println(err)
				fmt.Fprintf(w, "Erro ao criar log de erro da migration")
				return
			}
			fmt.Println(err)
			fmt.Fprintf(w, "Erro ao executar a migration")
			return
		}
	}
	fmt.Fprintf(w, "true")
}

// SaveMigration ... Salva nova migration
func SaveMigration(w http.ResponseWriter, r *http.Request) {
	_, err := models.InsertMigration(r.FormValue("title"), r.FormValue("query"))
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Erro ao executar insert")
		return
	}

	fmt.Fprintf(w, "true")
}

// DeleteMigration ... Exclui uma migratio que não tenha sido executada nem em teste e nem em produção
func DeleteMigration(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query()["id"][0])

	_, err := models.DeleteMigrationLog(id)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Erro ao excluir logs da migration")
		return
	}

	_, err2 := models.DeleteMigration(id)
	if err2 != nil {
		fmt.Println(err2)
		fmt.Fprintf(w, "Erro ao excluir migration")
		return
	}

	fmt.Fprintf(w, "true")
}

// ShowLogs ... Lista todos os logs da migrations selecionada
func ShowLogs(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query()["migrationID"][0])
	logs, err := models.GetMigrationLogs(id)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Erro ao excluir migration")
		return
	}

	jsonLogs, _ := json.Marshal(logs)

	fmt.Fprintf(w, string(jsonLogs))
}
