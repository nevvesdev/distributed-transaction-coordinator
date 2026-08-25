package redis

import (
	"context"
	"fmt"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/config"
	goredis "github.com/redis/go-redis/v9"
)

// NovaConexao cria e valida uma conexão com o Redis.
func NovaConexao(cfg config.ConfigRedis) (*goredis.Client, error) {
	cliente := goredis.NewClient(&goredis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Porta),
		Password:     cfg.Senha,
		DB:           cfg.DB,
		DialTimeout:  cfg.Timeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	ctx := context.Background()
	if err := cliente.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("erro ao conectar ao Redis: %w", err)
	}

	return cliente, nil
}
