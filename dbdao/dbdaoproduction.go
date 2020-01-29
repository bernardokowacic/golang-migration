package dbdao

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
)

func connProduction() (*sql.DB, error) {
	dbMigration := os.Getenv("DB_PRODUCTION")

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
