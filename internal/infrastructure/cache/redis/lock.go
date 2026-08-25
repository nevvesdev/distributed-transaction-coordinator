package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	// prefixoLock é o namespace das chaves de lock no Redis.
	prefixoLock = "dtc:lock:"
	// ttlLock é o tempo de vida padrão de um lock distribuído.
	ttlLock = 30 * time.Second
	// ttlRenovacao é o tempo adicionado ao renovar um lock.
	ttlRenovacao = 15 * time.Second
)

// ErrLockNaoAdquirido indica que o lock já está sendo mantido por outro processo.
var ErrLockNaoAdquirido = errors.New("lock não disponível: recurso em uso por outro processo")

// ErrLockNaoEncontrado indica que o lock não existe ao tentar liberar ou renovar.
var ErrLockNaoEncontrado = errors.New("lock não encontrado")

// RedisLock implementa lock.DistributedLock usando operações atômicas do Redis.
type RedisLock struct {
	cliente *goredis.Client
}

// NovoRedisLock cria uma nova instância do distributed lock.
func NovoRedisLock(cliente *goredis.Client) *RedisLock {
	return &RedisLock{cliente: cliente}
}

// Adquirir tenta adquirir o lock usando SET NX (atômico) com TTL.
// Retorna ErrLockNaoAdquirido se o lock já existir.
func (l *RedisLock) Adquirir(ctx context.Context, chave string) error {
	chaveCompleta := prefixoLock + chave

	adquirido, err := l.cliente.SetNX(ctx, chaveCompleta, "1", ttlLock).Result()
	if err != nil {
		return fmt.Errorf("erro ao tentar adquirir lock '%s': %w", chave, err)
	}
	if !adquirido {
		return ErrLockNaoAdquirido
	}

	return nil
}

// Liberar remove o lock do Redis para a chave informada.
func (l *RedisLock) Liberar(ctx context.Context, chave string) error {
	chaveCompleta := prefixoLock + chave

	removidos, err := l.cliente.Del(ctx, chaveCompleta).Result()
	if err != nil {
		return fmt.Errorf("erro ao liberar lock '%s': %w", chave, err)
	}
	if removidos == 0 {
		return ErrLockNaoEncontrado
	}

	return nil
}

func (l *RedisLock) Renovar(ctx context.Context, chave string) error {
	chaveCompleta := prefixoLock + chave

	renovado, err := l.cliente.Expire(ctx, chaveCompleta, ttlRenovacao).Result()
	if err != nil {
		return fmt.Errorf("erro ao renovar lock '%s': %w", chave, err)
	}
	if !renovado {
		return ErrLockNaoEncontrado
	}

	return nil
}
