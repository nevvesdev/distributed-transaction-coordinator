package middleware

import (
	"log"
	"net/http"
)

// Recuperacao é um middleware que captura panics e retorna 500
// sem derrubar o servidor.
func Recuperacao(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic capturado: %v — rota: %s %s", err, r.Method, r.URL.Path)
				http.Error(w, `{"erro":"erro interno do servidor"}`, http.StatusInternalServerError)
			}
		}()
		proximo.ServeHTTP(w, r)
	})
}
