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
	http.HandleFunc("/", index)
	http.HandleFunc("/update-production", updateProduction)
	http.HandleFunc("/update-test", updateTest)
	http.HandleFunc("/save-migration", saveMigration)
	http.ListenAndServe(":8000", nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	err := openDatabaseConnection()
	if err != nil {
		fmt.Println(err)
	}

	_, queries, err3 := listMigrations(0)
	if err3 != nil {
		fmt.Println(err3)
	}

	totalPagesFloat := float64(len(queries) / 15)
	totalPages := uint32(math.Ceil(totalPagesFloat))
	if totalPages <= 0 {
		totalPages = 1
	}

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

	closeDatabaseConnection()
}

func updateProduction(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "true")
}

func updateTest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "teste")
}

func saveMigration(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	err := conn.PingContext(ctx)
	if err != nil {
		//fmt.Fprintf(w, err)
		//return -1, err
	}

	tsql := `INSERT INTO migrations (
		name, query, created_at, executed_on_test, executed_on_production
	) VALUES (
		@name, @query, @created_at, @executed_on_test, @executed_on_production
	);`

	stmt, err := conn.Prepare(tsql)
	if err != nil {
		//fmt.Fprintf(w, err)
		//return -1, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(
		ctx,
		sql.Named("name", r.FormValue("title")),
		sql.Named("query", r.FormValue("query")),
		sql.Named("created_at", time.Now().Format("2006-01-02 15:04:05")),
		sql.Named("executed_on_test", 0),
		sql.Named("executed_on_test", 0),
	)
	var newID int64
	err = row.Scan(&newID)
	if err != nil {
		//fmt.Fprintf(w, err)
		//return -1, err
	}

	fmt.Fprintf(w, "true")
}

func openDatabaseConnection() error {
	connectionString := "sqlserver://sa:QWer1234*()@192.168.16.2:1433"
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
	switch filter {
	case 1: // Executado somente em produção
		query = query + " where executed_on_production = 1 and executed_on_test = 0"
	case 2: // Executado somente em teste
		query = query + " where executed_on_production = 0 and executed_on_test = 1"
	case 3: // Executado em teste e em produção
		query = query + " where executed_on_production = 1 and executed_on_test = 1"
	case 4: // Não executado em lugar nenhum
		query = query + " where executed_on_production = 0 and executed_on_test = 0"
	default:
		query = query
	}

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
