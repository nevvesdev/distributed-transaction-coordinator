package participant

import "context"

// Repository define o contrato de persistência para o agregado Participant.
type Repository interface {
	// Salvar persiste um novo participante.
	Salvar(ctx context.Context, participante *Participant) error
	// Atualizar persiste as mudanças de estado de um participante existente.
	Atualizar(ctx context.Context, participante *Participant) error
	// BuscarPorID retorna um participante pelo seu identificador único.
	BuscarPorID(ctx context.Context, id string) (*Participant, error)
	// ListarPorTransacao retorna todos os participantes de uma transação.
	ListarPorTransacao(ctx context.Context, idTransacao string) ([]*Participant, error)
	// DeletarPorTransacao remove todos os participantes de uma transação.
	DeletarPorTransacao(ctx context.Context, idTransacao string) error
}
