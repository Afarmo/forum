package database

import (
	"database/sql"
	"os"
)

func Open() (*sql.DB, bool, error) {
	_, err := os.Stat("forum.db")
	isNewDatabase := os.IsNotExist(err)
	if err != nil && !isNewDatabase {
		return nil, false, err
	}

	db, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		return nil, false, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, false, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, false, err
	}

	return db, isNewDatabase, nil
}

func InitializeSchema(db *sql.DB, seed bool) error {
	var schemaExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM sqlite_master
			WHERE type = 'table' AND name = 'USERS' COLLATE NOCASE
		) AND EXISTS (
			SELECT 1 FROM sqlite_master
			WHERE type = 'table' AND name = 'POSTS_SEARCH' COLLATE NOCASE
		)
	`).Scan(&schemaExists)
	if err != nil {
		return err
	}

	if !schemaExists {
		schemaBytes, err := os.ReadFile("internal/database/schema.sql")
		if err != nil {
			return err
		}

		if _, err := db.Exec(string(schemaBytes)); err != nil {
			return err
		}
	}

	if seed {
		if err := SeedCategories(db); err != nil {
			return err
		}
		if err := SeedDemoData(db); err != nil {
			return err
		}
	}

	return nil
}
