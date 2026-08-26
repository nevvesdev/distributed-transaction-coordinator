package saga

import "context"

// Repository define o contrato de persistência para o agregado Saga.
type Repository interface {
	// Salvar persiste uma nova Saga com seus steps.
	Salvar(ctx context.Context, saga *Saga) error
	// Atualizar persiste as mudanças de estado de uma Saga existente.
	Atualizar(ctx context.Context, saga *Saga) error
	// BuscarPorID retorna uma Saga pelo seu identificador único.
	BuscarPorID(ctx context.Context, id string) (*Saga, error)
	// BuscarPorTransacao retorna a Saga associada a uma transação.
	BuscarPorTransacao(ctx context.Context, idTransacao string) (*Saga, error)
}
