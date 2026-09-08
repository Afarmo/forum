package middleware

import (
	"log"
	"net/http"
	"time"
)

type StatusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *StatusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		sw := &StatusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)

		var color string
		switch {
		case sw.status >= 500:
			color = "\033[31m  [PANIC]\033[0m   "
		case sw.status >= 400:
			color = "\033[33m  [WARN]\033[0m    "
		default:
			color = "\033[32m  [LOG]\033[0m     "

		}

		log.Printf("%s %s %s %d %s",
			color,
			r.Method,
			r.URL.Path,
			sw.status,
			time.Since(start),
		)
	})
}
