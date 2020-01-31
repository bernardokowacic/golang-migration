package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migration/dbdao"
)

// Migrations ... Cria estrutura dos registros de migration
type Migrations struct {
	Codigo               uint
	Name                 string
	Query                string
	User                 sql.NullInt32
	CreatedAt            string
	ExecutedOnTest       bool
	ExecutedOnProduction bool
}

// GetAllMigrations ... Busca todas as migrations salvas no BD de migrations
func GetAllMigrations(filter uint16, page uint16) ([]Migrations, error) {
	query := "select * from migrations "

	pagination := fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT 15 ROWS ONLY", page)

	switch filter {
	case 1: // Executado somente em produção
		query = query + " where executed_on_test = 0"
	case 2: // Executado somente em teste
		query = query + " where executed_on_production = 0"
	}

	query = query + " order by created_at desc"
	query = query + pagination

	rows, err := dbdao.Select(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var migrations []Migrations
	for rows.Next() {
		var id uint
		var name string
		var query string
		var userID sql.NullInt32
		var createdAt string
		var executedOnTest bool
		var executedOnProduction bool

		// Get values from row.
		err := rows.Scan(&id, &name, &query, &userID, &createdAt, &executedOnTest, &executedOnProduction)
		if err != nil {
			return nil, nil
		}

		migrationDate, err := time.Parse("2006-01-02T15:04:05Z", createdAt)
		createdAt = migrationDate.Format("02/01/2006 15:04")

		migrations = append(migrations, Migrations{
			Codigo:               id,
			Name:                 name,
			Query:                query,
			User:                 userID,
			CreatedAt:            createdAt,
			ExecutedOnTest:       executedOnTest,
			ExecutedOnProduction: executedOnProduction,
		})
	}

	return migrations, nil
}

// ExecMigration ... Executa uma nova migration no ambiente de teste OU producao
func ExecMigration(migration Migrations, ambiente string) (bool, error) {
	if ambiente != "teste" && ambiente != "producao" {
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

		doUpdate = doUpdate + "executed_on_test = 1 where id in (?)"

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

		doUpdate = doUpdate + "executed_on_production = 1 where id in (?)"

		_, errUpdate := dbdao.ExecOnProductiont(doUpdate, migration.Codigo)

		if errUpdate != nil {
			fmt.Println("Error exec migration: ", errUpdate.Error())
			return false, errUpdate
		}
	}

	return true, nil
}

// InsertMigration ... Insere uma nova migration no BD de migrations
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

// GetMigrationsByID ... Busca as migrations selecionadas na hora de executar no BD de teste
func GetMigrationsByID(filter uint16, ids string) ([]Migrations, error) {
	query := "select * from migrations "

	switch filter {
	case 1: // Executado somente em produção
		query = query + " where executed_on_test = 0"
	case 2: // Executado somente em teste
		query = query + " where executed_on_production = 0"
	}
	query = query + " and id in (" + ids + ")"

	query = query + " order by created_at desc"

	rows, err := dbdao.Select(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var migrations []Migrations
	for rows.Next() {
		var id uint
		var name string
		var query string
		var userID sql.NullInt32
		var createdAt string
		var executedOnTest bool
		var executedOnProduction bool

		// Get values from row.
		err := rows.Scan(&id, &name, &query, &userID, &createdAt, &executedOnTest, &executedOnProduction)
		if err != nil {
			return nil, nil
		}

		migrationDate, err := time.Parse("2006-01-02T15:04:05Z", createdAt)
		createdAt = migrationDate.Format("02/01/2006 15:04")

		migrations = append(migrations, Migrations{
			Codigo:               id,
			Name:                 name,
			Query:                query,
			User:                 userID,
			CreatedAt:            createdAt,
			ExecutedOnTest:       executedOnTest,
			ExecutedOnProduction: executedOnProduction,
		})
	}

	return migrations, nil
}