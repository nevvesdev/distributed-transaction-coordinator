package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/shared/idempotency"
)

const cabecalhoIdempotencia = "Idempotency-Key"

// Idempotencia é um middleware que intercepta requisições com Idempotency-Key
// e retorna a resposta cached se a operação já foi processada anteriormente.
func Idempotencia(store idempotency.Store) func(http.Handler) http.Handler {
	return func(proximo http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// apenas métodos que modificam estado precisam de idempotência
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				proximo.ServeHTTP(w, r)
				return
			}

			chave := r.Header.Get(cabecalhoIdempotencia)
			if chave == "" {
				proximo.ServeHTTP(w, r)
				return
			}

			// verifica se já existe resposta cached para esta chave
			resultado, err := store.Buscar(r.Context(), chave)
			if err != nil {
				log.Printf("middleware idempotência — erro ao buscar chave '%s': %v", chave, err)
				proximo.ServeHTTP(w, r)
				return
			}

			// retorna resposta cached sem reprocessar
			if resultado != nil {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Idempotency-Replayed", "true")
				w.WriteHeader(resultado.StatusHTTP)
				_, _ = w.Write(resultado.Corpo)
				return
			}

			// captura a resposta para armazenar no cache
			capturador := &captorResposta{
				ResponseWriter: w,
				corpo:          &bytes.Buffer{},
				status:         http.StatusOK,
			}

			proximo.ServeHTTP(capturador, r)

			// persiste apenas respostas de sucesso
			if capturador.status < 400 {
				if err := store.Salvar(r.Context(), chave, idempotency.Resultado{
					StatusHTTP: capturador.status,
					Corpo:      capturador.corpo.Bytes(),
				}); err != nil {
					log.Printf("middleware idempotência — erro ao salvar chave '%s': %v", chave, err)
				}
			}
		})
	}
}

// captorResposta captura o corpo e status da resposta para armazenamento.
type captorResposta struct {
	http.ResponseWriter
	corpo  *bytes.Buffer
	status int
}

func (c *captorResposta) WriteHeader(status int) {
	c.status = status
	c.ResponseWriter.WriteHeader(status)
}

func (c *captorResposta) Write(b []byte) (int, error) {
	c.corpo.Write(b)
	return c.ResponseWriter.Write(b)
}

// responderJSON escreve uma resposta JSON com o status informado.
func responderJSON(w http.ResponseWriter, status int, corpo interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(corpo)
}

// lerCorpo lê e recoloca o corpo da requisição para uso posterior.
func lerCorpo(r *http.Request) ([]byte, error) {
	corpo, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(corpo))
	return corpo, nil
}
