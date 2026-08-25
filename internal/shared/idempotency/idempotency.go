package idempotency

import "context"

// Resultado representa a resposta cached de uma operação já processada.
type Resultado struct {
	StatusHTTP int
	Corpo      []byte
}

// Store define o contrato para armazenamento de chaves de idempotência.
// Garante que requisições duplicadas retornem a mesma resposta sem reprocessamento.
type Store interface {
	// Salvar persiste o resultado de uma operação associado à chave de idempotência.
	Salvar(ctx context.Context, chave string, resultado Resultado) error
	// Buscar retorna o resultado cached para a chave informada, se existir.
	Buscar(ctx context.Context, chave string) (*Resultado, error)
	// Existe verifica se a chave de idempotência já foi processada.
	Existe(ctx context.Context, chave string) (bool, error)
}
