package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logging é um middleware que registra método, rota, status e duração de cada requisição.
func Logging(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		proximo.ServeHTTP(rw, r)

		log.Printf("%s %s → %d (%s)",
			r.Method,
			r.URL.Path,
			rw.status,
			time.Since(inicio),
		)
	})
}

// responseWriter encapsula http.ResponseWriter para capturar o status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
