package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/golang-migration/models"
	"github.com/joho/godotenv"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/joho/godotenv"
)

func main() {
	errEnv := godotenv.Load()
	if errEnv != nil {
		log.Fatal("Error loading .env file")
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/update-production", updateProduction)
	http.HandleFunc("/update-test", updateTest)
	http.HandleFunc("/save-migration", saveMigration)
	http.ListenAndServe(":8000", nil)
}

var temp = template.Must(template.ParseGlob("templates/*.html"))

func index(w http.ResponseWriter, r *http.Request) {
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

func updateProduction(w http.ResponseWriter, r *http.Request) {
	queries, err := models.GetAllMigrations(2)
	if err != nil {
		fmt.Println(err)
	}

	if len(queries) < 1 {
		fmt.Println(err)
		fmt.Fprintf(w, "Todas as migrations já foram rodadas no BD de produção")
		return
	}

	db, err := sql.Open("mssql", dbMigration)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Conexão com o banco não está funcionando")
		return
	}
	prodConn = db
	ctxTest := context.Background()
	err = prodConn.PingContext(ctxTest)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Conexão com o banco de produção não está funcionando")
		return
	}
	defer prodConn.Close()

	for x := 0; x < len(queries); x++ {
		stmt, _ := prodConn.Prepare(queries[x].Query)
		_, err = stmt.Exec()
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Erro ao executar insert no BD de produção")
			return
		}

		prodStmt, _ := conn.Prepare("update migrations set Executed_on_production = 1 where id = ?")
		_, err := prodStmt.Exec(
			queries[x].Codigo,
		)
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Erro ao executar update na migration executada")
			return
		}
	}

	fmt.Fprintf(w, "true")
}

func updateTest(w http.ResponseWriter, r *http.Request) {
	var selectedMigrations []string //{ value string }
	err := json.Unmarshal([]byte(r.FormValue("migrationsToRun")), &selectedMigrations)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "erro ao receber queries à serem executadas")
		return
	}
	migrationsToRunIDs := strings.Join(selectedMigrations[:], ",")
	fmt.Println(migrationsToRunIDs) // FALTA AJUSTAR O SELECT PARA BUSCAR SOMENTE OS IDS QUE ESTÃO NESSA VARIÁVEL

	dbMigration := os.Getenv("DB_TEST")
	ctx := context.Background()
	err2 := conn.PingContext(ctx)
	if err2 != nil {
		fmt.Println(err2)
		fmt.Fprintf(w, "Conexão com o banco de migrations não está funcionando")
		return
	}
	_, queries, err3 := listMigrations(1)
	if err3 != nil {
		fmt.Println(err3)
	}

	if len(queries) < 1 {
		fmt.Println(err)
		fmt.Fprintf(w, "Todas as migrations já foram rodadas no BD de teste")
		return
	}

	db, err := sql.Open("mssql", dbMigration)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Conexão com o banco de teste não está funcionando")
		return
	}
	testConn = db
	ctxTest := context.Background()
	err = testConn.PingContext(ctxTest)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Conexão com o banco não está funcionando")
		return
	}
	defer testConn.Close()

	for x := 0; x < len(queries); x++ {
		stmt, _ := testConn.Prepare(queries[x].Query)
		_, err = stmt.Exec()
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Erro ao executar insert no BD de teste")
			return
		}

		testStmt, _ := conn.Prepare("update migrations set executed_on_test = 1 where id = ?")
		_, err := testStmt.Exec(
			queries[x].Codigo,
		)
		if err != nil {
			fmt.Println(err)
			fmt.Fprintf(w, "Erro ao executar update na migration executada")
			return
		}
	}
	fmt.Fprintf(w, "true")
}

func saveMigration(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	err := conn.PingContext(ctx)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Conexão com o banco não está funcionando")
		return
	}

	tsql :=
		`INSERT INTO migrations (
			name, query, created_at, executed_on_test, executed_on_production
		) VALUES (
			?, ?, ?, ?, ?
		);`

	stmt, _ := conn.Prepare(tsql)
	_, err = stmt.Exec(
		r.FormValue("title"),
		r.FormValue("query"),
		time.Now().Format("2006-01-02 15:04:05"),
		0,
		0,
	)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Erro ao executar insert")
		return
	}

	fmt.Fprintf(w, "true")
}

func openDatabaseConnection() error {
	dbMigration := os.Getenv("DB_MIGRATION")
	db, err := sql.Open("mssql", dbMigration)
	if err != nil {
		return err
	}
	conn = db
	ctx := context.Background()
	err = conn.PingContext(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	return nil
}
