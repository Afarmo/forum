package middleware

import (
	"html/template"
	"log"
	"net/http"

	"github.com/Afarmo/forum/internal/apperrors"
)

func Recover(next http.Handler, tmpl *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("\033[31m  [PANIC]\033[0m    %s %s %v", r.Method, r.URL.Path, err)
				apperrors.RenderError(w, tmpl, http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
