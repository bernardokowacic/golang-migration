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
	SQL   dbdao.QueryPaginate
	Items []ItemsMigration
}

type ItemsMigration struct {
	Codigo               uint
	Name                 string
	Query                string
	User                 sql.NullInt32
	CreatedAt            string
	ExecutedOnTest       bool
	ExecutedOnProduction bool
}

// GetAllMigrations ... Busca todas as migrations salvas no BD de migrations
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
		var userID sql.NullInt32
		var createdAt string
		var executedOnTest bool
		var executedOnProduction bool

		// Get values from row.
		err := appendResult.SQL.Rows.Scan(&id, &name, &query, &userID, &createdAt, &executedOnTest, &executedOnProduction)
		if err != nil {
			return Migrations{}, err
		}

		migrationDate, err := time.Parse("2006-01-02T15:04:05Z", createdAt)
		if err != nil {
			return Migrations{}, err
		}
		createdAt = migrationDate.Format("02/01/2006 15:04")

		appendResult.Items = append(appendResult.Items, ItemsMigration{
			Codigo:               id,
			Name:                 name,
			Query:                query,
			User:                 userID,
			CreatedAt:            createdAt,
			ExecutedOnTest:       executedOnTest,
			ExecutedOnProduction: executedOnProduction,
		})
	}

	return appendResult, nil
}

// ExecMigration ... Executa uma nova migration no ambiente de teste OU producao
func ExecMigration(migration ItemsMigration, ambiente string) (bool, error) {
	if ambiente != "teste" && ambiente != "producao" {
		return false, errors.New("o parametro ambiente deve ser 'teste' ou 'producao'")
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
func GetMigrationsByID(filter uint16, ids string) (Migrations, error) {
	var objQuery dbdao.ReceivedQuery
	var appendResult Migrations
	objQuery.Select = "select * from migrations"
	var err error

	// pagination := fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT 15 ROWS ONLY", page)

	switch filter {
	case 1: // Executado somente em produção
		objQuery.Where = "where executed_on_test = 0 and id in (" + ids + ")"
	case 2: // Executado somente em teste
		objQuery.Where = "where executed_on_production = 0 and id in (" + ids + ")"
	}

	objQuery.Order = "order by executed_on_production asc, executed_on_test asc, created_at desc"
	// query = query + pagination

	appendResult.SQL, err = dbdao.Select(objQuery, 0)
	if err != nil {
		return Migrations{}, err
	}

	defer appendResult.SQL.Rows.Close()

	for appendResult.SQL.Rows.Next() {
		var id uint
		var name string
		var query string
		var userID sql.NullInt32
		var createdAt string
		var executedOnTest bool
		var executedOnProduction bool

		// Get values from row.
		err := appendResult.SQL.Rows.Scan(&id, &name, &query, &userID, &createdAt, &executedOnTest, &executedOnProduction)
		if err != nil {
			return Migrations{}, err
		}

		migrationDate, err := time.Parse("2006-01-02T15:04:05Z", createdAt)
		if err != nil {
			return Migrations{}, err
		}
		createdAt = migrationDate.Format("02/01/2006 15:04")

		appendResult.Items = append(appendResult.Items, ItemsMigration{
			Codigo:               id,
			Name:                 name,
			Query:                query,
			User:                 userID,
			CreatedAt:            createdAt,
			ExecutedOnTest:       executedOnTest,
			ExecutedOnProduction: executedOnProduction,
		})
	}

	return appendResult, nil
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
