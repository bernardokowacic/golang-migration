package models

import (
	"fmt"
	"time"

	"github.com/golang-migration/dbdao"
)

// Logs ... Cria estrutura dos registros de log
type Logs struct {
	ID          uint
	Description string
	CreatedAt   string
}

// GetMigrationLogs ... Retorna todos os logs de uma migration
func GetMigrationLogs(migrationID int) ([]Logs, error) {
	var objQuery dbdao.ReceivedQuery

	objQuery.Select = fmt.Sprintf("select id, description, created_at from logs where migration_id = %d", migrationID)

	rows, err := dbdao.Select(objQuery, -1)
	if err != nil {
		return nil, err
	}

	defer rows.Rows.Close()

	var logs []Logs
	for rows.Rows.Next() {
		var id uint
		var description string
		var createdAt string

		err := rows.Rows.Scan(&id, &description, &createdAt)
		if err != nil {
			return nil, err
		}

		logDate, err := time.Parse("2006-01-02T15:04:05Z", createdAt)
		if err != nil {
			return nil, err
		}
		createdAt = logDate.Format("02/01/2006 15:04:05")

		logs = append(logs, Logs{
			ID:          id,
			Description: description,
			CreatedAt:   createdAt,
		})
	}

	return logs, nil
}

// DeleteMigrationLog ... Deleta todos os logs de uma migration
func DeleteMigrationLog(migrationID int) (bool, error) {
	query := "delete from logs where migration_id = ?"

	_, err := dbdao.ExecOnMigration(query, migrationID)
	if err != nil {
		return false, err
	}

	return true, nil
}

// CreateMigrationLog ... Insere novo log de uma determinada migration
func CreateMigrationLog(migrationID uint, logText string) (bool, error) {
	query := "insert into logs (migration_id, description, created_at) values (?, ?, ?)"

	_, err := dbdao.ExecOnMigration(query, migrationID, logText, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return false, err
	}

	return true, nil
}
