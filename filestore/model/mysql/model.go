package mysql

import (
	"database/sql"
	"log"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// InitMysql -
func InitMysql() (writer *sql.DB, reader *sql.DB) {
	writer, err := sql.Open("mysql", "root:123456@tcp(localhost:3307)/fileserver")
	if err != nil {
		log.Fatalln(err)
	}

	writer.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	reader, err = sql.Open("mysql", "reader:123456@tcp(localhost:3308)/fileserver")
	if err != nil {
		log.Fatalln(err)
	}

	reader.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	writer.SetMaxOpenConns(1000)
	reader.SetMaxOpenConns(10000)

	err = createFileTable(writer)
	if err != nil {
		log.Fatalln(err)
	}

	return writer, reader
}
