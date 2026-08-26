package dlq

import "context"

// Repository define o contrato de persistência para mensagens da DLQ.
type Repository interface {
	// Salvar persiste uma nova mensagem na DLQ.
	Salvar(ctx context.Context, mensagem *Mensagem) error
	// Atualizar persiste as mudanças de estado de uma mensagem existente.
	Atualizar(ctx context.Context, mensagem *Mensagem) error
	// BuscarPorID retorna uma mensagem pelo seu identificador único.
	BuscarPorID(ctx context.Context, id string) (*Mensagem, error)
	// ListarPendentes retorna mensagens prontas para reprocessamento.
	ListarPendentes(ctx context.Context, limite int) ([]*Mensagem, error)
	// ListarPorReferencia retorna todas as mensagens de uma referência.
	ListarPorReferencia(ctx context.Context, idReferencia string) ([]*Mensagem, error)
}
