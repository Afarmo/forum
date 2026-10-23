package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestSearchPostsWithoutFTSTable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL,
			profile_picture TEXT
		);
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			picture_content TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		INSERT INTO users (id, username, profile_picture) VALUES (1, 'alice', '');
		INSERT INTO posts (user_id, title, content, picture_content, created_at, updated_at)
		VALUES (1, 'Hello world', 'This is a test post', '', ?, ?);
	` , time.Now(), time.Now())
	if err != nil {
		t.Fatalf("init schema: %v", err)
	}

	repo := NewPostRepository(db)
	posts, err := repo.SearchPosts(context.Background(), "hello")
	if err != nil {
		t.Fatalf("SearchPosts returned error: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if posts[0].Title != "Hello world" {
		t.Fatalf("expected matching title, got %q", posts[0].Title)
	}
}
