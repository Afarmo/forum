package app

import (
	"database/sql"
	"html/template"
)

// Application holds the shared dependencies handlers and middleware need
// (DB connection, parsed templates, and anything added later such as a
// session store or logger), so they can be passed around explicitly
// instead of relying on package-level globals.
type Application struct {
	DB       *sql.DB
	Template *template.Template
}
