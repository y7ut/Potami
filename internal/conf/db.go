package conf

import (
	"database/sql"
	"log"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var (
	sqliteClient   *sql.DB
	SqliteInitOnce sync.Once
)

func SqliteDB() *sql.DB {
	SqliteInitOnce.Do(func() {
		var err error
		sqliteClient, err = sql.Open(DB.Type, DB.Path)
		if err != nil {
			log.Fatal(err)
		}
	})
	return sqliteClient
}
