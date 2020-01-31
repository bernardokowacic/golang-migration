package dbdao

import (
	"database/sql"
	"fmt"
	"os"

	// Esta aqui pq o go adiciona sozinho
	_ "github.com/denisenkom/go-mssqldb"
)

func conn() (*sql.DB, error) {
	dbMigration := os.Getenv("DB_MIGRATION")

	db, err := sql.Open("mssql", dbMigration)
	if err != nil {
		fmt.Println("Cannot connect: ", err.Error())
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Cannot connect: ", err.Error())
		return nil, err
	}
	return db, nil
}

// Select ... Executa queries de select
func Select(query string, where ...interface{}) (*sql.Rows, error) {
	db, err := conn()
	defer db.Close()

	if err != nil {
		fmt.Println("Error select: ", err.Error())
		return nil, err
	}

	rows, rowsErr := db.Query(query, where...)

	if rowsErr != nil {
		fmt.Println("Error select: ", rowsErr.Error())
		return nil, rowsErr
	}

	return rows, nil
}

// ExecOnMigration ... Executa queries geradas
func ExecOnMigration(query string, args ...interface{}) (bool, error) {
	db, err := conn()
	defer db.Close()

	if err != nil {
		fmt.Println("Error exec migration: ", err.Error())
		return false, err
	}

	stmt, _ := db.Prepare(query)
	_, errStmt := stmt.Exec(args...)

	if errStmt != nil {
		fmt.Println("Error exec migration: ", errStmt.Error())
		return false, errStmt
	}

	return true, nil
}
