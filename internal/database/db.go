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
func BackfillPostSearch(db *sql.DB) error {
	_, err := db.Exec(`
        INSERT INTO POSTS_SEARCH(rowid, title, content, username)
        SELECT POSTS.id, POSTS.title, POSTS.content, USERS.username
        FROM POSTS
        JOIN USERS ON USERS.id = POSTS.user_id
        WHERE NOT EXISTS (
            SELECT 1
            FROM POSTS_SEARCH
            WHERE POSTS_SEARCH.rowid = POSTS.id
        )
    `)
	return err
}

func BackfillPostSearch(db *sql.DB) error {
	_, err := db.Exec(`
        INSERT INTO POSTS_SEARCH(rowid, title, content, username)
        SELECT POSTS.id, POSTS.title, POSTS.content, USERS.username
        FROM POSTS
        JOIN USERS ON USERS.id = POSTS.user_id
        WHERE NOT EXISTS (
            SELECT 1
            FROM POSTS_SEARCH
            WHERE POSTS_SEARCH.rowid = POSTS.id
        )
    `)
	return err
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
	if err := BackfillPostSearch(db); err != nil {
		return err
	}
	err = SeedCategories(db)
	if err != nil {
		return err
	}

	err = SeedDemoData(db)
	if err != nil {
		return err
	}

	return nil
}
