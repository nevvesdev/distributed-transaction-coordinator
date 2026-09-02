package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/query"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/transaction"
)

// AuditHandler expõe os endpoints de consulta do audit trail e projeção de eventos.
type AuditHandler struct {
	handlerConsulta *query.HandlerConsulta
}

// NovoAuditHandler cria uma nova instância do handler de auditoria.
func NovoAuditHandler(handlerConsulta *query.HandlerConsulta) *AuditHandler {
	return &AuditHandler{handlerConsulta: handlerConsulta}
}

// AuditTrail retorna o histórico completo de eventos de um agregado.
// GET /audit/{id_agregado}
func (h *AuditHandler) AuditTrail(w http.ResponseWriter, r *http.Request) {
	idAgregado := chi.URLParam(r, "id_agregado")

	resultado, err := h.handlerConsulta.ConsultarAuditTrail(r.Context(), query.ConsultarAuditTrail{
		IDAgregado: idAgregado,
	})
	if err != nil {
		ResponderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	ResponderJSON(w, http.StatusOK, resultado)
}

// ProjecaoTransacao reconstrói o histórico de transições de estado via Event Sourcing.
// GET /audit/transacoes/{id}/projecao
func (h *AuditHandler) ProjecaoTransacao(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	projecao, err := h.handlerConsulta.ProjetarEstadoTransacao(r.Context(), id)
	if err != nil {
		ResponderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	ResponderJSON(w, http.StatusOK, projecao)
}

// DetalheTransacao retorna o estado atual de uma transação com formatação de consulta.
// GET /audit/transacoes/{id}
func (h *AuditHandler) DetalheTransacao(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	resultado, err := h.handlerConsulta.ConsultarTransacao(r.Context(), query.ConsultarTransacao{
		ID: id,
	})
	if err != nil {
		if errors.Is(err, transaction.ErrTransacaoNaoEncontrada) {
			ResponderErro(w, http.StatusNotFound, "transação não encontrada")
			return
		}
		ResponderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	ResponderJSON(w, http.StatusOK, resultado)
}

// ParticipantesTransacao retorna os participantes formatados de uma transação.
// GET /audit/transacoes/{id}/participantes
func (h *AuditHandler) ParticipantesTransacao(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	participantes, err := h.handlerConsulta.ConsultarParticipantes(r.Context(), id)
	if err != nil {
		ResponderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	ResponderJSON(w, http.StatusOK, participantes)
}
