package database

import (
	"database/sql"
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
