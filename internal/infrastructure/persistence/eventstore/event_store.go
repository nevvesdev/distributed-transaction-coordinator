package eventstore

import (
	"context"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/shared/domain"
)

// EventStore define o contrato para persistência de eventos de domínio.
type EventStore interface {
	// Salvar persiste uma lista de eventos de domínio.
	Salvar(ctx context.Context, eventos []domain.DomainEvent) error
	// ListarPorAgregado retorna todos os eventos de um agregado ordenados por ocorrência.
	ListarPorAgregado(ctx context.Context, idAgregado string) ([]RegistroEvento, error)
}

// RegistroEvento representa um evento persistido no banco de dados.
type RegistroEvento struct {
	ID         int64
	IDAgregado string
	NomeEvento string
	Payload    []byte
	OcorridoEm string
}
