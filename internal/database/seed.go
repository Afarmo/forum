package database

import (
	"database/sql"
	"strings"

	"github.com/Afarmo/forum/internal/auth"
)

func SeedCategories(db *sql.DB) error {
	categories := []string{
		"Fantasy", "Romance", "Mystery", "Thriller", "Horror",
		"Science Fiction", "Adventure", "Historical Fiction",
		"Literary Fiction", "Contemporary Fiction", "Crime", "Drama",
		"Young Adult", "Children's", "Biography", "Autobiography",
		"Memoir", "History", "Poetry", "Classics", "Dystopian",
		"Self-Help", "Philosophy", "Psychology", "Graphic Novel",
	}

	query := `INSERT OR IGNORE INTO categories (name) VALUES (?)`
	for _, category := range categories {
		if _, err := db.Exec(query, category); err != nil {
			return err
		}
	}
	return nil
}

// SeedDemoData populates the database with demo users, posts and comments.
func SeedDemoData(db *sql.DB) error {
	var postCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM posts`).Scan(&postCount); err != nil {
		return err
	}
	if postCount > 0 {
		return nil
	}

	const password = "password123"

	users := []struct {
		username string
		email    string
	}{
		{"alice", "alice@example.com"},
		{"bob", "bob@example.com"},
		{"carol", "carol@example.com"},
		{"dave", "dave@example.com"},
		{"eve", "eve@example.com"},
	}

	userIDs := make(map[string]int, len(users))
	for _, u := range users {
		id, err := ensureSeedUser(db, u.username, u.email, password)
		if err != nil {
			return err
		}
		userIDs[u.username] = id
	}

	categoryID := func(name string) (int, error) {
		var id int
		err := db.QueryRow(`SELECT id FROM categories WHERE name = ?`, name).Scan(&id)
		return id, err
	}

	posts := []struct {
		author   string
		category string
		title    string
		content  string
	}{
		{"alice", "Fantasy", "Why The Name of the Wind still holds up",
			"A decade later, Rothfuss's prose still reads like a love letter to storytelling itself. The magic system feels grounded, but the real hook is Kvothe's voice. What did you think of the framing device?"},
		{"bob", "Mystery", "Cozy mysteries for rainy weekends",
			"I've been building a list of low-stakes whodunits with endearing detectives. Agatha Christie's Miss Marple is obvious, but the Thursday Murder Club books are great modern comfort reads."},
		{"carol", "Classics", "Re-reading Pride and Prejudice: social satire",
			"It's easy to read this as pure romance, but Charlotte Lucas alone justifies the whole novel as economic commentary. Which character's motivation landed differently for you on a second read?"},
		{"dave", "Science Fiction", "Dune vs. Foundation: worldbuilding",
			"Herbert and Asimov both build empires across millennia, but Herbert's ecology-as-politics focus feels more relevant now. Which future empire felt more believable to you?"},
		{"eve", "Young Adult", "YA picks that resonate with adults",
			"YA often gets dismissed, but stories like The Hate U Give carry weight any adult should read. What are your crossover favorites?"},
		{"alice", "Horror", "Gothic horror recommendations beyond Dracula",
			"Shirley Jackson's Hill House is the obvious starter, but I'd add Rebecca and The Little Stranger. The slow dread beats jump scares every time."},
		{"bob", "Crime", "True crime that reads like fiction",
			"In Cold Blood and Killers of the Flower Moon both read like thrillers while documenting real tragedy. How do we stay respectful as readers?"},
		{"carol", "Historical Fiction", "Historical fiction beyond the Tudors",
			"The Tudor well is deep, but Pachinko, All the Light We Cannot See, and Homegoing stretch the genre beautifully. What eras deserve more novels?"},
		{"dave", "Memoir", "Memoirs that changed how I see failure",
			"Educated and Born a Crime both reframe hardship as a kind of education. Which memoir made you reconsider your own setbacks?"},
		{"eve", "Poetry", "Poetry for people who think they hate poetry",
			"Start with Mary Oliver or Billy Collins—clear images, no decoder ring required. What poem surprised you into caring?"},
	}

	postIDs := make(map[string]int, len(posts))
	for _, p := range posts {
		cid, err := categoryID(p.category)
		if err != nil {
			return err
		}
		res, err := db.Exec(
			`INSERT INTO posts (user_id, title, content, picture_content) VALUES (?,?,?,?)`,
			userIDs[p.author], p.title, p.content, "",
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		postID := int(id)
		postIDs[p.title] = postID
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES (?,?)`,
			postID, cid,
		); err != nil {
			return err
		}
	}

	comments := []struct {
		post    string
		author  string
		content string
		parent  int
	}{
		{"Why The Name of the Wind still holds up", "bob", "The framing device is the best part—Kvothe telling his own myth is irresistible.", -1},
		{"Why The Name of the Wind still holds up", "carol", "Agreed. The Chandrian mystery keeps the middle from sagging for me.", -1},
		{"Why The Name of the Wind still holds up", "alice", "Thanks! The Chandrian tease is exactly what kept me hooked too.", 1},
		{"Cozy mysteries for rainy weekends", "alice", "Adding The Windsor Knot to my list—great suggestion avenue.", -1},
		{"Cozy mysteries for rainy weekends", "dave", "Miss Marple forever. The Mirror Crack'd from Side to Side is underrated.", -1},
		{"Cozy mysteries for rainy weekends", "bob", "Mirror Crack'd is fantastic; glad someone mentioned it.", 4},
		{"Re-reading Pride and Prejudice: social satire", "dave", "Charlotte Lucas is the most pragmatic character in all of Austen.", -1},
		{"Re-reading Pride and Prejudice: social satire", "eve", "Lydia's storyline hits harder when you realize how little agency she actually had.", -1},
		{"Re-reading Pride and Prejudice: social satire", "carol", "Yes—Charlotte's bargain reads differently once you've lived a little.", 6},
		{"Dune vs. Foundation: worldbuilding", "alice", "The Bene Gesserit ecology angle is what makes Dune timeless.", -1},
		{"Dune vs. Foundation: worldbuilding", "eve", "Foundation's psychohistory is cooler as an idea, but Dune's worlds feel lived-in.", -1},
		{"Dune vs. Foundation: worldbuilding", "dave", "Both are engineering feats, but Dune wins on atmosphere.", 9},
		{"YA picks that resonate with adults", "bob", "The Hate U Give made me cry on a plane. No shame.", -1},
		{"YA picks that resonate with adults", "carol", "Same. Also loved They Both Die at the End for its tenderness.", -1},
		{"YA picks that resonate with adults", "eve", "Adding both to my TBR. Thanks!", 13},
		{"Gothic horror recommendations beyond Dracula", "dave", "The Haunting of Hill House is top tier. The opening paragraph alone.", -1},
		{"Gothic horror recommendations beyond Dracula", "bob", "Rebecca is perfect—so much dread without a single ghost.", -1},
		{"Gothic horror recommendations beyond Dracula", "alice", "Exactly—psychological dread over spectacle, always.", 15},
		{"True crime that reads like fiction", "carol", "Killers of the Flower Moon wrecked me. The scope of injustice was staggering.", -1},
		{"True crime that reads like fiction", "eve", "Same. I had to put it down twice.", -1},
		{"True crime that reads like fiction", "bob", "Putting it down is fair—pace yourself.", 19},
		{"Historical fiction beyond the Tudors", "alice", "Pachinko is a masterpiece. Four generations and not one wasted page.", -1},
		{"Historical fiction beyond the Tudors", "dave", "All the Light We Cannot See haunts me still.", -1},
		{"Historical fiction beyond the Tudors", "carol", "Agreed on both—broadening the genre matters.", 22},
		{"Memoirs that changed how I see failure", "eve", "Educated reshaped how I think about education and family.", -1},
		{"Memoirs that changed how I see failure", "bob", "Born a Crime's humor makes the pain land even harder.", -1},
		{"Memoirs that changed how I see failure", "dave", "Amen to both. Storytelling as survival.", 24},
		{"Poetry for people who think they hate poetry", "alice", "Mary Oliver's Wild Geese changed how I walk in the woods.", -1},
		{"Poetry for people who think they hate poetry", "carol", "Billy Collins is the gateway drug. Highly recommend Sailing Alone Around the Room.", -1},
		{"Poetry for people who think they hate poetry", "eve", "Thanks—adding Oliver immediately.", 27},
	}

	commentIDs := make([]int, len(comments))
	for i, c := range comments {
		var parent any
		if c.parent >= 0 {
			parent = commentIDs[c.parent]
		}
		res, err := db.Exec(
			`INSERT INTO comments (user_id, post_id, parent_id, content) VALUES (?,?,?,?)`,
			userIDs[c.author], postIDs[c.post], parent, c.content,
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		commentIDs[i] = int(id)
	}

	return nil
}

func ensureSeedUser(db *sql.DB, username, email, password string) (int, error) {
	var id int
	err := db.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	hash, salt, err := auth.HashPassword(password)
	if err != nil {
		return 0, err
	}

	res, err := db.Exec(
		`INSERT INTO users (username, email, password_hash, password_salt, profile_picture) VALUES (?,?,?,?,?)`,
		username, email, hash, salt, "",
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if err2 := db.QueryRow(`SELECT id FROM users WHERE username = ?`, username).Scan(&id); err2 == nil {
				return id, nil
			}
		}
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(lastID), nil
}
