package dbdao

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/denisenkom/go-mssqldb"
)

type funcQueries interface {
	totalPages() int64
}

type QueryPaginate struct {
	Rows       *sql.Rows
	Page       uint64
	PerPage    uint64
	TotalItems uint64
}

type ReceivedQuery struct {
	Select   string
	Order    string
	Where    string
	Paginate string
}

type funcReceivedQuery interface {
	queryToSelect() int64
}

func (r *ReceivedQuery) queryToSelect() string {
	return fmt.Sprintf("%s %s %s %s", r.Select, r.Where, r.Order, r.Paginate)
}

func (q *QueryPaginate) totalPages() uint64 {
	return uint64(q.TotalItems / q.PerPage)
}

func conn() (*sql.DB, error) {
	dbMigration := os.Getenv("DB_MIGRATION")

	db, err := sql.Open("mssql", dbMigration)
	if err != nil {
		log.Println("Cannot connect: ", err.Error())
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		log.Println("Cannot connect: ", err.Error())
		return nil, err
	}
	return db, nil
}

func Select(query ReceivedQuery, page int) (QueryPaginate, error) {
	var dataQuery QueryPaginate
	var totalValues uint64
	db, err := conn()

	if err != nil {
		log.Println("Error select: ", err.Error())
		return dataQuery, err
	}

	defer db.Close()

	if page != -1 {
		rowsCount, err := db.Query(fmt.Sprintf("select count(*) from (%s) as sq ", query.Select))

		if err != nil {
			log.Println(err)
		}

		defer rowsCount.Close()

		rowsCount.Next()

		rowsCount.Scan(&totalValues)

		query.Paginate = fmt.Sprintf("OFFSET %d ROWS FETCH NEXT 15 ROWS ONLY", page)

		dataQuery.TotalItems = totalValues
	}

	rows, rowsErr := db.Query(query.queryToSelect())

	if rowsErr != nil {
		log.Println("Error select: ", rowsErr.Error())
		return dataQuery, rowsErr
	}

	dataQuery.Rows = rows

	return dataQuery, nil
}

func ExecOnMigration(query string, args ...interface{}) (bool, error) {
	db, err := conn()
	defer db.Close()

	if err != nil {
		log.Println("Error exec migration: ", err.Error())
		return false, err
	}

	stmt, _ := db.Prepare(query)
	_, errStmt := stmt.Exec(args...)

	if errStmt != nil {
		log.Println("Error exec migration: ", errStmt.Error())
		return false, errStmt
	}

	return true, nil
}
