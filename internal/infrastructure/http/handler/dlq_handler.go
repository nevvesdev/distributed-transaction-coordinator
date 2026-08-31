package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"unsafe"

	"github.com/go-chi/chi/v5"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/service"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/dlq"
)

// DLQHandler expõe os endpoints de gerenciamento da Dead Letter Queue.
type DLQHandler struct {
	worker *service.WorkerDLQ
}

// NovoDLQHandler cria uma nova instância do handler.
func NovoDLQHandler(worker *service.WorkerDLQ) *DLQHandler {
	return &DLQHandler{worker: worker}
}

// Enfileirar adiciona manualmente uma mensagem na DLQ.
// POST /dlq
func (h *DLQHandler) Enfileirar(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDReferencia string          `json:"id_referencia"`
		Tipo         string          `json:"tipo"`
		Payload      json.RawMessage `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if req.IDReferencia == "" || req.Tipo == "" {
		responderErro(w, http.StatusBadRequest, "id_referencia e tipo são obrigatórios")
		return
	}

	if err := h.worker.EnfileirarMensagem(r.Context(), req.IDReferencia, req.Tipo, req.Payload); err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, map[string]string{
		"mensagem": "mensagem enfileirada na DLQ com sucesso",
	})
}

// ListarPorReferencia retorna as mensagens DLQ de uma referência.
// GET /dlq/{id_referencia}
func (h *DLQHandler) ListarPorReferencia(w http.ResponseWriter, r *http.Request) {
	idReferencia := chi.URLParam(r, "id_referencia")

	mensagens, err := listarDLQPorReferencia(r.Context(), h.worker, idReferencia)
	if err != nil {
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, mensagens)
}

func listarDLQPorReferencia(ctx context.Context, worker *service.WorkerDLQ, idReferencia string) ([]*dlq.Mensagem, error) {
	field := reflect.ValueOf(worker).Elem().FieldByName("repoDLQ")
	if !field.IsValid() {
		return nil, fmt.Errorf("campo repoDLQ não encontrado")
	}

	campoAcessivel := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	metodo := campoAcessivel.MethodByName("ListarPorReferencia")
	if !metodo.IsValid() {
		return nil, fmt.Errorf("método ListarPorReferencia não encontrado no repositório")
	}

	resultados := metodo.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(idReferencia)})
	if len(resultados) != 2 {
		return nil, fmt.Errorf("assinatura inesperada do repositório DLQ")
	}
	if !resultados[1].IsNil() {
		return nil, resultados[1].Interface().(error)
	}

	mensagens, ok := resultados[0].Interface().([]*dlq.Mensagem)
	if !ok {
		return nil, fmt.Errorf("tipo de retorno inesperado para ListarPorReferencia")
	}
	return mensagens, nil
}

// responderJSON escreve uma resposta JSON — compartilhado entre handlers.
func responderJSON(w http.ResponseWriter, status int, corpo interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(corpo)
}

// ResponderJSON escreve uma resposta JSON para a camada HTTP.
func ResponderJSON(w http.ResponseWriter, status int, corpo interface{}) {
	responderJSON(w, status, corpo)
}

// responderErro escreve uma resposta de erro padronizada.
func responderErro(w http.ResponseWriter, status int, mensagem string) {
	responderJSON(w, status, map[string]string{"erro": mensagem})
}

// ResponderErro escreve uma resposta de erro padronizada.
func ResponderErro(w http.ResponseWriter, status int, mensagem string) {
	responderErro(w, status, mensagem)
}
