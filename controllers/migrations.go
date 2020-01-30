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
	queries, err := models.GetAllMigrations(0)
	if err != nil {
		fmt.Println(err)
	}

	totalPagesFloat := float64(len(queries) / 15)
	totalPages := uint32(math.Ceil(totalPagesFloat))
	if totalPages <= 0 {
		totalPages = 1
	}

	fmt.Println(totalPages)

	totalRegisters := 2

	var currentPage uint32 = 1

	transformJSON, _ := json.Marshal(queries)

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
		"Queries":         queries,
		"Json":            string(transformJSON),
	}

	err2 := temp.ExecuteTemplate(w, "Index", data)
	if err2 != nil {
		fmt.Println(err2)
	}
}

func UpdateProduction(w http.ResponseWriter, r *http.Request) {
	queries, err := models.GetAllMigrations(2)
	if err != nil {
		fmt.Println(err)
	}

	if len(queries) < 1 {
		fmt.Println(err)
		fmt.Fprintf(w, "Todas as migrations já foram rodadas no BD de produção")
		return
	}

	// for x := 0; x < len(queries); x++ {
	// 	stmt, _ := prodConn.Prepare(queries[x].Query)
	// 	_, err = stmt.Exec()
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		fmt.Fprintf(w, "Erro ao executar insert no BD de produção")
	// 		return
	// 	}

	// 	prodStmt, _ := conn.Prepare("update migrations set Executed_on_production = 1 where id = ?")
	// 	_, err := prodStmt.Exec(
	// 		queries[x].Codigo,
	// 	)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		fmt.Fprintf(w, "Erro ao executar update na migration executada")
	// 		return
	// 	}
	// }

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

	queries, err := models.GetAllMigrations(2)
	if err != nil {
		fmt.Println(err)
	}

	if len(queries) < 1 {
		fmt.Println(err)
		fmt.Fprintf(w, "Todas as migrations já foram rodadas no BD de teste")
		return
	}

	// for x := 0; x < len(queries); x++ {
	// 	stmt, err := dbdao.ExecMigration(queries[x])
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		fmt.Fprintf(w, "Erro ao executar update na migration executada")
	// 		return
	// 	}
	// }
	fmt.Fprintf(w, "true")
}

func SaveMigration(w http.ResponseWriter, r *http.Request) {
	// tsql :=
	// 	`INSERT INTO migrations (
	// 		name, query, created_at, executed_on_test, executed_on_production
	// 	) VALUES (
	// 		?, ?, ?, ?, ?
	// 	);`

	// stmt, _ := conn.Prepare(tsql)
	// _, err = stmt.Exec(
	// 	r.FormValue("title"),
	// 	r.FormValue("query"),
	// 	time.Now().Format("2006-01-02 15:04:05"),
	// 	0,
	// 	0,
	// )
	// if err != nil {
	// 	fmt.Println(err)
	// 	fmt.Fprintf(w, "Erro ao executar insert")
	// 	return
	// }

	fmt.Fprintf(w, "true")
}
