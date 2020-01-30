package dbdao

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
)

func connTest() (*sql.DB, error) {
	dbMigration := os.Getenv("DB_TEST")

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

func ExecOnTest(query string, args ...interface{}) (bool, error) {
	db, err := connTest()
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
