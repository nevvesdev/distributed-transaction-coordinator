package http

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/http/handler"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/http/middleware"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/observability"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/shared/idempotency"
)

// ConfigurarRotas registra todas as rotas da aplicação no roteador chi.
func ConfigurarRotas(
	transacaoHandler *handler.TransacaoHandler,
	participanteHandler *handler.ParticipanteHandler,
	sagaHandler *handler.SagaHandler,
	dlqHandler *handler.DLQHandler,
	auditHandler *handler.AuditHandler,
	healthHandler *handler.HealthHandler,
	idemStore idempotency.Store,
	metricas *observability.Metricas,
) *chi.Mux {
	r := chi.NewRouter()

	// middlewares globais
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.Recuperacao)
	r.Use(middleware.Logging)
	r.Use(middleware.Metricas(metricas))
	r.Use(middleware.Idempotencia(idemStore))

	// health check detalhado
	r.Get("/health", healthHandler.Verificar)

	// métricas Prometheus
	r.Handle("/metrics", observability.Handler())

	// rotas de transações 2PC
	r.Route("/transacoes", func(r chi.Router) {
		r.Post("/", transacaoHandler.Criar)
		r.Get("/{id}", transacaoHandler.Buscar)
		r.Post("/{id}/prepare", transacaoHandler.IniciarPrepare)
		r.Post("/{id}/commit", transacaoHandler.IniciarCommit)
		r.Post("/{id}/abort", transacaoHandler.Abortar)
		r.Post("/{id}/participantes", participanteHandler.Registrar)
		r.Get("/{id}/participantes", participanteHandler.Listar)
	})

	// rotas de saga orquestrada
	r.Route("/sagas", func(r chi.Router) {
		r.Post("/", sagaHandler.Executar)
		r.Get("/{id}", sagaHandler.Buscar)
	})

	// rotas da DLQ
	r.Route("/dlq", func(r chi.Router) {
		r.Post("/", dlqHandler.Enfileirar)
		r.Get("/{id_referencia}", dlqHandler.ListarPorReferencia)
	})

	// rotas de auditoria e event sourcing
	r.Route("/audit", func(r chi.Router) {
		r.Get("/{id_agregado}", auditHandler.AuditTrail)
		r.Get("/transacoes/{id}", auditHandler.DetalheTransacao)
		r.Get("/transacoes/{id}/projecao", auditHandler.ProjecaoTransacao)
		r.Get("/transacoes/{id}/participantes", auditHandler.ParticipantesTransacao)
	})

	return r
}
