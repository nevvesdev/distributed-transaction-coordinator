package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/command"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/service"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/transaction"
)

// TransacaoHandler expõe os endpoints do ciclo de vida da transação 2PC.
type TransacaoHandler struct {
	coordinador *service.Coordinador2PC
}

// NovoTransacaoHandler cria uma nova instância do handler.
func NovoTransacaoHandler(coordinador *service.Coordinador2PC) *TransacaoHandler {
	return &TransacaoHandler{coordinador: coordinador}
}

// requisicaoIniciarTransacao é o corpo esperado para criar uma transação.
type requisicaoIniciarTransacao struct {
	Payload   map[string]string `json:"payload"`
	Timeout   string            `json:"timeout"`
	ChaveIdem string            `json:"chave_idempotencia"`
}

// Criar cria uma nova transação distribuída.
// POST /transacoes
func (h *TransacaoHandler) Criar(w http.ResponseWriter, r *http.Request) {
	var req requisicaoIniciarTransacao
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if req.ChaveIdem == "" {
		responderErro(w, http.StatusBadRequest, "chave_idempotencia é obrigatória")
		return
	}

	timeout, err := time.ParseDuration(req.Timeout)
	if err != nil || timeout <= 0 {
		timeout = 30 * time.Second
	}

	resultado, err := h.coordinador.IniciarTransacao(r.Context(), command.IniciarTransacao{
		Payload:   req.Payload,
		Timeout:   timeout,
		ChaveIdem: req.ChaveIdem,
	})
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, resultado)
}

// Buscar retorna uma transação pelo ID.
// GET /transacoes/{id}
func (h *TransacaoHandler) Buscar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		responderErro(w, http.StatusBadRequest, "id da transação é obrigatório")
		return
	}

	transacao, err := h.coordinador.BuscarTransacao(r.Context(), id)
	if err != nil {
		if errors.Is(err, transaction.ErrTransacaoNaoEncontrada) {
			responderErro(w, http.StatusNotFound, "transação não encontrada")
			return
		}
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, transacao)
}

// IniciarPrepare inicia a fase de prepare do 2PC.
// POST /transacoes/{id}/prepare
func (h *TransacaoHandler) IniciarPrepare(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	resultado, err := h.coordinador.ProcessarPrepare(r.Context(), command.ProcessarPrepare{
		IDTransacao: id,
	})
	if err != nil {
		if errors.Is(err, transaction.ErrTransacaoExpirada) {
			responderErro(w, http.StatusGone, "transação expirada")
			return
		}
		responderErro(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, resultado)
}

// IniciarCommit inicia a fase de commit do 2PC.
// POST /transacoes/{id}/commit
func (h *TransacaoHandler) IniciarCommit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	resultado, err := h.coordinador.ProcessarCommit(r.Context(), command.ProcessarCommit{
		IDTransacao: id,
	})
	if err != nil {
		responderErro(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, resultado)
}

// Abortar reverte uma transação.
// POST /transacoes/{id}/abort
func (h *TransacaoHandler) Abortar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Motivo string `json:"motivo"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Motivo == "" {
		req.Motivo = "abort solicitado pelo cliente"
	}

	resultado, err := h.coordinador.AbortarTransacao(r.Context(), command.AbortarTransacao{
		IDTransacao: id,
		Motivo:      req.Motivo,
	})
	if err != nil {
		responderErro(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, resultado)
}
