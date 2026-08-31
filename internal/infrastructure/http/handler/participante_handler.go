package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/command"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/service"
)

// ParticipanteHandler expõe os endpoints de gerenciamento de participantes.
type ParticipanteHandler struct {
	coordinador *service.Coordinador2PC
}

// NovoParticipanteHandler cria uma nova instância do handler.
func NovoParticipanteHandler(coordinador *service.Coordinador2PC) *ParticipanteHandler {
	return &ParticipanteHandler{coordinador: coordinador}
}

// requisicaoRegistrarParticipante é o corpo esperado para registrar um participante.
type requisicaoRegistrarParticipante struct {
	Endpoint string `json:"endpoint"`
}

// Registrar adiciona um participante a uma transação existente.
// POST /transacoes/{id}/participantes
func (h *ParticipanteHandler) Registrar(w http.ResponseWriter, r *http.Request) {
	idTransacao := chi.URLParam(r, "id")

	var req requisicaoRegistrarParticipante
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if req.Endpoint == "" {
		responderErro(w, http.StatusBadRequest, "endpoint do participante é obrigatório")
		return
	}

	resultado, err := h.coordinador.RegistrarParticipante(r.Context(), command.RegistrarParticipante{
		IDTransacao: idTransacao,
		Endpoint:    req.Endpoint,
	})
	if err != nil {
		responderErro(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, resultado)
}

// Listar retorna todos os participantes de uma transação.
// GET /transacoes/{id}/participantes
func (h *ParticipanteHandler) Listar(w http.ResponseWriter, r *http.Request) {
	idTransacao := chi.URLParam(r, "id")

	participantes, err := h.coordinador.ListarParticipantes(r.Context(), idTransacao)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, participantes)
}
