package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migration/dbdao"
)

type Migrations struct {
	SQL   dbdao.QueryPaginate
	Items []ItemsMigration
}

type ItemsMigration struct {
	Codigo                 uint
	Name                   string
	Query                  string
	User                   sql.NullInt32
	Created_at             string
	Executed_on_test       bool
	Executed_on_production bool
}

func GetAllMigrations(filter uint16, page int) (Migrations, error) {
	var objQuery dbdao.ReceivedQuery
	var appendResult Migrations
	objQuery.Select = "select * from migrations"
	var err error

	// pagination := fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT 15 ROWS ONLY", page)

	switch filter {
	case 1: // Executado somente em produção
		objQuery.Where = "where executed_on_test = 0"
	case 2: // Executado somente em teste
		objQuery.Where = "where executed_on_production = 0"
	}

	objQuery.Order = "order by executed_on_production asc, executed_on_test asc, created_at desc"
	// query = query + pagination

	appendResult.SQL, err = dbdao.Select(objQuery, page)
	if err != nil {
		return Migrations{}, err
	}

	defer appendResult.SQL.Rows.Close()

	for appendResult.SQL.Rows.Next() {
		var id uint
		var name string
		var query string
		var user_id sql.NullInt32
		var created_at string
		var executed_on_test bool
		var executed_on_production bool

		// Get values from row.
		err := appendResult.SQL.Rows.Scan(&id, &name, &query, &user_id, &created_at, &executed_on_test, &executed_on_production)
		if err != nil {
			return Migrations{}, nil
		}

		migrationDate, err := time.Parse("2006-01-02T15:04:05Z", created_at)
		created_at = migrationDate.Format("02/01/2006 15:04")

		appendResult.Items = append(appendResult.Items, ItemsMigration{
			Codigo:                 id,
			Name:                   name,
			Query:                  query,
			User:                   user_id,
			Created_at:             created_at,
			Executed_on_test:       executed_on_test,
			Executed_on_production: executed_on_production,
		})
	}

	return appendResult, nil
}

func ExecMigration(migration ItemsMigration, ambiente string) (bool, error) {
	if ambiente != "teste" || ambiente != "producao" {
		return false, errors.New("o parametro ambiente deve ser teste ou producao")
	}

	doUpdate := "update migrations set "

	switch ambiente {
	case "teste":
		_, errMigration := dbdao.ExecOnTest(migration.Query)

		if errMigration != nil {
			fmt.Println("Error exec migration: ", errMigration.Error())
			return false, errMigration
		}

		doUpdate = doUpdate + "executed_on_test = 1 where id = ?"

		_, errUpdate := dbdao.ExecOnTest(doUpdate, migration.Codigo)

		if errUpdate != nil {
			fmt.Println("Error exec migration: ", errUpdate.Error())
			return false, errUpdate
		}

	case "producao":
		_, errMigration := dbdao.ExecOnProductiont(migration.Query)

		if errMigration != nil {
			fmt.Println("Error exec migration: ", errMigration.Error())
			return false, errMigration
		}

		doUpdate = doUpdate + "executed_on_production = 1 where id = ?"

		_, errUpdate := dbdao.ExecOnProductiont(doUpdate, migration.Codigo)

		if errUpdate != nil {
			fmt.Println("Error exec migration: ", errUpdate.Error())
			return false, errUpdate
		}
	}

	return true, nil
}

func InsertMigration(title string, query string) (bool, error) {
	tsql :=
		`INSERT INTO migrations (
			name, query, created_at, executed_on_test, executed_on_production
		) VALUES (
			?, ?, ?, ?, ?
		);`

	_, err := dbdao.ExecOnMigration(tsql, title, query, time.Now().Format("2006-01-02 15:04:05"), 0, 0)

	if err != nil {
		return false, err
	}

	return true, nil
}
