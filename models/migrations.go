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

// Columns ... Cria estrutura das colunas do BD
type Columns struct {
	Name string
}

// Tables ... Cria estrutura das tabelas do BD
type Tables struct {
	Name    string
	Columns []Columns
}

// // Tables ... Cria estrutura das tabelas do BD
// type Tables map[string]Columns

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

		doUpdate = doUpdate + "executed_on_test = 1 where id = ?"

		_, errUpdate := dbdao.ExecOnMigration(doUpdate, migration.Codigo)
		if errUpdate != nil {
			fmt.Println("Error updating migration: ", errUpdate.Error())
			return false, errUpdate
		}

	case "producao":
		_, errMigration := dbdao.ExecOnProduction(migration.Query)

		if errMigration != nil {
			fmt.Println("Error exec migration: ", errMigration.Error())
			return false, errMigration
		}

		doUpdate = doUpdate + "executed_on_production = 1 where id = ?"

		_, errUpdate := dbdao.ExecOnMigration(doUpdate, migration.Codigo)

		if errUpdate != nil {
			fmt.Println("Error updating migration: ", errUpdate.Error())
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

// DeleteMigration ... Executa query que exclui 1 registro do banco
func DeleteMigration(id int) (bool, error) {
	tsql := `DELETE FROM migrations WHERE id = ? and executed_on_test = 0 and executed_on_production = 0;`

	_, err := dbdao.ExecOnMigration(tsql, id)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ShowAllColumns .. Lista todas as colunas do BD
func ShowAllColumns() (map[string]interface{}, error) {

	query := "select tab.name as table_name, col.name as column_name from sys.tables as tab inner join sys.columns as col on tab.object_id = col.object_id order by table_name asc, column_name asc"

	rows, err := dbdao.SelectOnTest(query)
	if err != nil {
		return nil, err
	}

	returnData := make(map[string]interface{})

	for rows.Next() {
		var table string
		var column string

		// Get values from row.
		err := rows.Scan(&table, &column)
		if err != nil {
			return nil, nil
		}

		if returnData[table] != nil {
			returnData[table] = append(returnData[table].([]string), column)
		} else {
			returnData[table] = []string{column}
		}
	}

	return returnData, nil
}
