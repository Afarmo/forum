package apperrors

import (
	"bytes"
	"html/template"
	"net/http"
)

func RenderError(w http.ResponseWriter, tmpl *template.Template, status int) {
	var buf bytes.Buffer

	if err := tmpl.ExecuteTemplate(&buf, "errors.html", status); err != nil {
		Log(err)
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}
