package database

import "database/sql"

func SeedCategories(db *sql.DB) error {

	categories := []string{
		"Fantasy",
		"Romance",
		"Mystery",
		"Thriller",
		"Horror",
		"Science Fiction",
		"Adventure",
		"Historical Fiction",
		"Literary Fiction",
		"Contemporary Fiction",
		"Crime",
		"Drama",
		"Young Adult",
		"Children's",
		"Biography",
		"Autobiography",
		"Memoir",
		"History",
		"Poetry",
		"Classics",
		"Dystopian",
		"Self-Help",
		"Philosophy",
		"Psychology",
		"Graphic Novel",
	}

	query := `
        INSERT OR IGNORE INTO categories (name)
        VALUES (?)
    `
	for _, category := range categories {
		_, err := db.Exec(query, category)

		if err != nil {
			return err
		}
	}

	return nil
}
