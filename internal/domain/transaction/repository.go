package transaction

import "context"

// Repository define o contrato de persistência para o agregado Transaction.
// A implementação concreta vive na camada de infraestrutura.
type Repository interface {
	// Salvar persiste uma nova transação.
	Salvar(ctx context.Context, transacao *Transaction) error
	// Atualizar persiste as mudanças de estado de uma transação existente.
	Atualizar(ctx context.Context, transacao *Transaction) error
	// BuscarPorID retorna uma transação pelo seu identificador único.
	BuscarPorID(ctx context.Context, id string) (*Transaction, error)
	// BuscarPorChaveIdem retorna uma transação pela chave de idempotência.
	BuscarPorChaveIdem(ctx context.Context, chave string) (*Transaction, error)
	// ListarExpiradas retorna transações que ultrapassaram o tempo limite.
	ListarExpiradas(ctx context.Context) ([]*Transaction, error)
}
