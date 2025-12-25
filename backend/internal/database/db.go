package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}

func ApplySchema(db *sql.DB, schemaPath string) {
	b, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(string(b)); err != nil {
		log.Fatal(err)
	}
}
