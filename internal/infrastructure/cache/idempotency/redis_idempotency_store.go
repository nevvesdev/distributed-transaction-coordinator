package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/shared/idempotency"
	goredis "github.com/redis/go-redis/v9"
)

const (
	// prefixoIdem é o namespace das chaves de idempotência no Redis.
	prefixoIdem = "dtc:idem:"
	// ttlIdem é o tempo de retenção de uma chave de idempotência.
	ttlIdem = 24 * time.Hour
)

// RedisIdempotencyStore implementa idempotency.Store usando Redis.
type RedisIdempotencyStore struct {
	cliente *goredis.Client
}

// NovoRedisIdempotencyStore cria uma nova instância do store de idempotência.
func NovoRedisIdempotencyStore(cliente *goredis.Client) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{cliente: cliente}
}

// Salvar persiste o resultado de uma operação no Redis com TTL de 24 horas.
func (s *RedisIdempotencyStore) Salvar(ctx context.Context, chave string, resultado idempotency.Resultado) error {
	dados, err := json.Marshal(resultado)
	if err != nil {
		return fmt.Errorf("erro ao serializar resultado de idempotência: %w", err)
	}

	chaveCompleta := prefixoIdem + chave
	if err := s.cliente.Set(ctx, chaveCompleta, dados, ttlIdem).Err(); err != nil {
		return fmt.Errorf("erro ao salvar chave de idempotência '%s': %w", chave, err)
	}

	return nil
}

// Buscar retorna o resultado cached para a chave informada, se existir.
func (s *RedisIdempotencyStore) Buscar(ctx context.Context, chave string) (*idempotency.Resultado, error) {
	chaveCompleta := prefixoIdem + chave

	dados, err := s.cliente.Get(ctx, chaveCompleta).Bytes()
	if errors.Is(err, goredis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar chave de idempotência '%s': %w", chave, err)
	}

	var resultado idempotency.Resultado
	if err := json.Unmarshal(dados, &resultado); err != nil {
		return nil, fmt.Errorf("erro ao desserializar resultado de idempotência: %w", err)
	}

	return &resultado, nil
}

// Existe verifica se a chave de idempotência já foi processada.
func (s *RedisIdempotencyStore) Existe(ctx context.Context, chave string) (bool, error) {
	chaveCompleta := prefixoIdem + chave

	contagem, err := s.cliente.Exists(ctx, chaveCompleta).Result()
	if err != nil {
		return false, fmt.Errorf("erro ao verificar chave de idempotência '%s': %w", chave, err)
	}

	return contagem > 0, nil
}
