package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"text/template"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

var conn *sql.DB
var testConn *sql.DB
var prodConn *sql.DB
var temp = template.Must(template.ParseGlob("templates/*.html"))

type Migrations struct {
	Codigo                 uint
	Name                   string
	Query                  string
	User                   sql.NullInt32
	Created_at             string
	Executed_on_test       bool
	Executed_on_production bool
}

func main() {
	err := openDatabaseConnection()
	if err != nil {
		fmt.Println(err)
	}

	http.HandleFunc("/", index)
	http.HandleFunc("/update-production", updateProduction)
	http.HandleFunc("/update-test", updateTest)
	http.HandleFunc("/save-migration", saveMigration)
	http.ListenAndServe(":8000", nil)

	closeDatabaseConnection()
}

func index(w http.ResponseWriter, r *http.Request) {
	_, queries, err3 := listMigrations(0)
	if err3 != nil {
		fmt.Println(err3)
	}

	fmt.Println(len(queries))
	totalPagesFloat := float64(len(queries) / 15)
	fmt.Println(totalPagesFloat)
	totalPages := uint32(math.Ceil(totalPagesFloat))
	fmt.Println(totalPages)
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
	ctx := context.Background()
	err := conn.PingContext(ctx)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "Conexão com o banco de migrations não está funcionando")
		return
	}
	_, queries, err3 := listMigrations(2)
	if err3 != nil {
		fmt.Println(err3)
	}

	if len(queries) < 1 {
		fmt.Println(err)
		fmt.Fprintf(w, "Todas as migrations já foram rodadas no BD de produção")
		return
	}

	connectionString := "sqlserver://sa:QWer1234*()@192.168.16.2:1435"
	db, err := sql.Open("mssql", connectionString)
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
	ctx := context.Background()
	err := conn.PingContext(ctx)
	if err != nil {
		fmt.Println(err)
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

	connectionString := "sqlserver://sa:db11%23@192.168.50.24:1433"
	db, err := sql.Open("mssql", connectionString)
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
	connectionString := "sqlserver://sa:db11%23@192.168.50.24:1433?database=MORPHEUS_MIGRATIONS"
	db, err := sql.Open("mssql", connectionString)
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

func closeDatabaseConnection() {
	conn.Close()
}

func listMigrations(filter uint8) (int, []Migrations, error) {
	ctx := context.Background()
	err := conn.PingContext(ctx)
	if err != nil {
		return -1, nil, err
	}

	query := "select * from migrations"
	pagination := ""
	switch filter {
	case 1: // Executado somente em produção
		query = query + " where executed_on_test = 0"
	case 2: // Executado somente em teste
		query = query + " where executed_on_production = 0"
	default:
		query = query
		pagination = " OFFSET 0 ROWS FETCH NEXT 15 ROWS ONLY"
	}
	query = query + " order by created_at desc"
	query = query + pagination

	fmt.Println(query)

	tsql := fmt.Sprintf(query)
	rows, err := conn.Query(tsql)
	if err != nil {
		fmt.Println(err)
		return -1, nil, err
	}
	defer rows.Close()

	var count int
	var migrations []Migrations
	for rows.Next() {
		var id uint
		var name string
		var query string
		var user_id sql.NullInt32
		var created_at string
		var executed_on_test bool
		var executed_on_production bool

		// Get values from row.
		err := rows.Scan(&id, &name, &query, &user_id, &created_at, &executed_on_test, &executed_on_production)
		if err != nil {
			return -1, nil, err
		}

		migrationDate, err := time.Parse("2006-01-02T15:04:05Z", created_at)
		created_at = migrationDate.Format("02/01/2006 15:04")

		migrations = append(migrations, Migrations{
			Codigo:                 id,
			Name:                   name,
			Query:                  query,
			User:                   user_id,
			Created_at:             created_at,
			Executed_on_test:       executed_on_test,
			Executed_on_production: executed_on_production,
		})

		count++
	}

	return 1, migrations, nil
}
