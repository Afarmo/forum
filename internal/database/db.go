package database

import (
	"database/sql"
	"os"
)

func Open() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func InitializeSchema(db *sql.DB) error {
	schemaBytes, err := os.ReadFile("internal/database/schema.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(schemaBytes))

	if err != nil {
		return err
	}
	err = SeedCategories(db)
	if err != nil {
		return err
	}
	
	return err
}
