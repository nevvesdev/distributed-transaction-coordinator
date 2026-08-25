package lock

import "context"

// DistributedLock define o contrato para aquisição e liberação de locks distribuídos.
// Garante que apenas um processo execute uma operação crítica por vez.
type DistributedLock interface {
	// Adquirir tenta adquirir o lock para a chave informada.
	// Retorna erro se o lock já estiver sendo mantido por outro processo.
	Adquirir(ctx context.Context, chave string) error
	// Liberar libera o lock para a chave informada.
	Liberar(ctx context.Context, chave string) error
	// Renovar estende o TTL do lock para evitar expiração prematura.
	Renovar(ctx context.Context, chave string) error
}
