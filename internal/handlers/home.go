package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
)

func HomeHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "404 Not Found", http.StatusNotFound)
		}

		var buf bytes.Buffer

		if err := tmpl.ExecuteTemplate(&buf, "index.html", nil); err != nil {
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
