package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/observability"
)

// Metricas é um middleware que registra duração e contagem das requisições HTTP.
func Metricas(metricas *observability.Metricas) func(http.Handler) http.Handler {
	return func(proximo http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inicio := time.Now()

			metricas.OperacoesAtivas.Inc()
			defer metricas.OperacoesAtivas.Dec()

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			proximo.ServeHTTP(rw, r)

			metricas.DuracaoRequisicao.WithLabelValues(
				r.Method,
				r.URL.Path,
				strconv.Itoa(rw.status),
			).Observe(time.Since(inicio).Seconds())
		})
	}
}
