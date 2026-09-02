package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/query"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/service"
	infraidempotency "github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/cache/idempotency"
	infraredis "github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/cache/redis"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/config"
	infrahttp "github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/http"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/http/handler"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/observability"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/persistence/eventstore"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/persistence/mysql"
)

func main() {
	fmt.Println("Distributed Transaction Coordinator — inicializando...")

	cfg, err := config.Carregar()
	if err != nil {
		log.Fatalf("erro ao carregar configurações: %v", err)
	}

	// infraestrutura de dados
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
	repoSaga := mysql.NovoRepositorioSaga(db)
	repoDLQ := mysql.NovoRepositorioDLQ(db)

	// infraestrutura compartilhada
	eventoStore := eventstore.NovoMySQLEventStore(db)
	lockDistribuido := infraredis.NovoRedisLock(redisCliente)
	idemStore := infraidempotency.NovoRedisIdempotencyStore(redisCliente)
	metricas := observability.NovasMetricas()

	// application services
	coordinador := service.NovoCoordinador2PC(
		repoTransacao,
		repoParticipante,
		eventoStore,
		lockDistribuido,
		cfg.Transacao.TimeoutPrepare,
	)

	orquestrador := service.NovoOrchestradorSaga(
		repoSaga,
		eventoStore,
		lockDistribuido,
	)

	workerDLQ := service.NovoWorkerDLQ(
		repoDLQ,
		coordinador,
		orquestrador,
		15*time.Second,
		cfg.Transacao.IntervaloBaseRetry,
		cfg.Transacao.MaxTentativasDLQ,
	)

	// query handler — lado de leitura do CQRS
	handlerConsulta := query.NovoHandlerConsulta(
		repoTransacao,
		repoParticipante,
		eventoStore,
	)

	// workers em background
	ctx, cancelar := context.WithCancel(context.Background())
	defer cancelar()

	service.NovoMonitorTimeout(repoTransacao, coordinador, 10*time.Second).Iniciar(ctx)
	workerDLQ.Iniciar(ctx)

	// handlers HTTP
	transacaoHandler := handler.NovoTransacaoHandler(coordinador)
	participanteHandler := handler.NovoParticipanteHandler(coordinador)
	sagaHandler := handler.NovoSagaHandler(orquestrador)
	dlqHandler := handler.NovoDLQHandler(workerDLQ)
	auditHandler := handler.NovoAuditHandler(handlerConsulta)
	healthHandler := handler.NovoHealthHandler(db, redisCliente)

	// roteador e servidor
	router := infrahttp.ConfigurarRotas(
		transacaoHandler,
		participanteHandler,
		sagaHandler,
		dlqHandler,
		auditHandler,
		healthHandler,
		idemStore,
		metricas,
	)

	servidor := infrahttp.NovoServidor(cfg.Servidor, router)
	servidor.Iniciar()

	fmt.Println("✅ MySQL conectado")
	fmt.Println("✅ Redis conectado")
	fmt.Println("✅ Coordinador 2PC pronto")
	fmt.Println("✅ Orquestrador Saga pronto")
	fmt.Println("✅ Monitor de timeout ativo")
	fmt.Println("✅ Worker DLQ ativo")
	fmt.Println("✅ Event Sourcing e Audit Trail prontos")
	fmt.Println("✅ Métricas Prometheus em /metrics")
	fmt.Printf("✅ servidor HTTP na porta %s\n", cfg.Servidor.Porta)

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("sinal de encerramento recebido...")
	cancelar()

	ctxDesligar, cancelarDesligar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelarDesligar()

	if err := servidor.Desligar(ctxDesligar); err != nil {
		log.Fatalf("erro ao encerrar servidor: %v", err)
	}

	log.Println("servidor encerrado com sucesso")
}
