package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"unsafe"

	"github.com/go-chi/chi/v5"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/service"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/saga"
)

// SagaHandler expõe os endpoints de gerenciamento da Saga Orquestrada.
type SagaHandler struct {
	orquestrador *service.OrchestradorSaga
}

// NovoSagaHandler cria uma nova instância do handler.
func NovoSagaHandler(orquestrador *service.OrchestradorSaga) *SagaHandler {
	return &SagaHandler{orquestrador: orquestrador}
}

// requisicaoDefinicaoStep é o corpo de um step para criação da Saga.
type requisicaoDefinicaoStep struct {
	Nome           string `json:"nome"`
	Ordem          int    `json:"ordem"`
	Endpoint       string `json:"endpoint"`
	EndpointCompen string `json:"endpoint_compensacao"`
}

// requisicaoIniciarSaga é o corpo esperado para executar uma Saga.
type requisicaoIniciarSaga struct {
	IDTransacao string                    `json:"id_transacao"`
	Nome        string                    `json:"nome"`
	Steps       []requisicaoDefinicaoStep `json:"steps"`
}

// Executar cria e executa uma nova Saga Orquestrada.
// POST /sagas
func (h *SagaHandler) Executar(w http.ResponseWriter, r *http.Request) {
	var req requisicaoIniciarSaga
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderErro(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if req.IDTransacao == "" {
		responderErro(w, http.StatusBadRequest, "id_transacao é obrigatório")
		return
	}
	if req.Nome == "" {
		responderErro(w, http.StatusBadRequest, "nome da saga é obrigatório")
		return
	}
	if len(req.Steps) == 0 {
		responderErro(w, http.StatusBadRequest, "ao menos um step é obrigatório")
		return
	}

	steps := make([]service.DefinicaoStep, 0, len(req.Steps))
	for _, s := range req.Steps {
		steps = append(steps, service.DefinicaoStep{
			Nome:           s.Nome,
			Ordem:          s.Ordem,
			Endpoint:       s.Endpoint,
			EndpointCompen: s.EndpointCompen,
		})
	}

	resultado, err := h.orquestrador.Executar(r.Context(), service.ComandoIniciarSaga{
		IDTransacao: req.IDTransacao,
		NomeSaga:    req.Nome,
		Steps:       steps,
	})
	if err != nil {
		responderErro(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, resultado)
}

// Buscar retorna uma Saga pelo ID.
// GET /sagas/{id}
func (h *SagaHandler) Buscar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s, err := buscarSagaPorID(r.Context(), h.orquestrador, id)
	if err != nil {
		if errors.Is(err, saga.ErrSagaNaoEncontrada) {
			responderErro(w, http.StatusNotFound, "saga não encontrada")
			return
		}
		responderErro(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, s)
}

func buscarSagaPorID(ctx context.Context, orquestrador *service.OrchestradorSaga, id string) (*saga.Saga, error) {
	field := reflect.ValueOf(orquestrador).Elem().FieldByName("repoSaga")
	if !field.IsValid() {
		return nil, fmt.Errorf("campo repoSaga não encontrado")
	}

	campoAcessivel := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
	metodo := campoAcessivel.MethodByName("BuscarPorID")
	if !metodo.IsValid() {
		return nil, fmt.Errorf("método BuscarPorID não encontrado no repositório")
	}

	resultados := metodo.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(id)})
	if len(resultados) != 2 {
		return nil, fmt.Errorf("assinatura inesperada do repositório de saga")
	}
	if !resultados[1].IsNil() {
		return nil, resultados[1].Interface().(error)
	}
	if resultados[0].IsNil() {
		return nil, saga.ErrSagaNaoEncontrada
	}

	s, ok := resultados[0].Interface().(*saga.Saga)
	if !ok {
		return nil, fmt.Errorf("tipo de retorno inesperado para BuscarPorID")
	}
	return s, nil
}
