package models

import (
	"database/sql"
	"time"

	"github.com/golang-migration/dbdao"
)

type Migrations struct {
	Codigo                 uint
	Name                   string
	Query                  string
	User                   sql.NullInt32
	Created_at             string
	Executed_on_test       bool
	Executed_on_production bool
}

func GetAllMigrations(filter uint16) ([]Migrations, error) {
	query := "select * from migrations"
	pagination := " "
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
		var user_id sql.NullInt32
		var created_at string
		var executed_on_test bool
		var executed_on_production bool

		// Get values from row.
		err := rows.Scan(&id, &name, &query, &user_id, &created_at, &executed_on_test, &executed_on_production)
		if err != nil {
			return nil, nil
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
	}

	return migrations, nil
}
