package dbdao

import (
	"database/sql"
	"fmt"
	"os"

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

func Select(query string) (*sql.Rows, error) {
	db, err := conn()
	defer db.Close()

	if err != nil {
		if err != nil {
			fmt.Println("Error select: ", err.Error())
			return nil, err
		}
	}

	rows, err := db.Query(query)
	if err != nil {
		if err != nil {
			fmt.Println("Error select: ", err.Error())
			return nil, err
		}
	}

	return rows, nil
}
