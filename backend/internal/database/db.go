package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func OpenDB() *sql.DB {
	dsn := os.Getenv("DATABASE_URL")

	var db *sql.DB
	var err error

	if dsn == "" {
		db, err = sql.Open("sqlite3", "./forum.db?_foreign_keys=on")
	} else {
		db, err = sql.Open("postgres", dsn)
	}

	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}

func ApplySchema(db *sql.DB) {
	var schema string
	if os.Getenv("DATABASE_URL") == "" {
		schema = "schema.sql"
	} else {
		schema = "schema_pg.sql"
	}

	b, err := os.ReadFile(schema)
	if err != nil {
		log.Fatal(err)
	}

	if _, err = db.Exec(string(b)); err != nil {
		log.Fatal(err)
	}
}
