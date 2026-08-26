package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/service"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/cache/idempotency"
	infraredis "github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/cache/redis"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/config"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/persistence/eventstore"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/persistence/mysql"
)

func main() {
	fmt.Println("Distributed Transaction Coordinator — inicializando...")

	cfg, err := config.Carregar()
	if err != nil {
		log.Fatalf("erro ao carregar configurações: %v", err)
	}

	db, err := mysql.NovaConexao(cfg.MySQL)
	if err != nil {
		log.Fatalf("erro ao conectar ao MySQL: %v", err)
	}
	defer db.Close()

	if err := mysql.ExecutarMigracoes(db); err != nil {
		log.Fatalf("erro ao executar migrações: %v", err)
	}

	redisCliente, err := infraredis.NovaConexao(cfg.Redis)
	if err != nil {
		log.Fatalf("erro ao conectar ao Redis: %v", err)
	}
	defer redisCliente.Close()

	// repositórios
	repoTransacao := mysql.NovoRepositorioTransacao(db)
	repoParticipante := mysql.NovoRepositorioParticipante(db)

	// infraestrutura
	eventoStore := eventstore.NovoMySQLEventStore(db)
	lockDistribuido := infraredis.NovoRedisLock(redisCliente)
	_ = idempotency.NovoRedisIdempotencyStore(redisCliente)

	// application services
	coordinador := service.NovoCoordinador2PC(
		repoTransacao,
		repoParticipante,
		eventoStore,
		lockDistribuido,
		cfg.Transacao.TimeoutPrepare,
	)

	// monitor de timeout em background
	monitor := service.NovoMonitorTimeout(repoTransacao, coordinador, 10*time.Second)
	monitor.Iniciar(context.Background())

	fmt.Println("✅ MySQL conectado")
	fmt.Println("✅ Redis conectado")
	fmt.Println("✅ Coordinador 2PC pronto")
	fmt.Println("✅ Monitor de timeout ativo")
	fmt.Printf("✅ servidor pronto na porta %s\n", cfg.Servidor.Porta)
}
