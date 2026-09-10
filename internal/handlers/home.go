package handlers

import (
	"bytes"
	"log"
	"net/http"

	"github.com/Afarmo/forum/internal/app"
)

func HomeHandler(a *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "404 Not Found", http.StatusNotFound)
			return
		}

		var buf bytes.Buffer

		data := map[string]string{"Title": "Home"}

		if err := a.Template.ExecuteTemplate(&buf, "layout.html", data); err != nil {
			log.Println(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if _, err := buf.WriteTo(w); err != nil {
			log.Println(err)
			return
		}
	}
}
