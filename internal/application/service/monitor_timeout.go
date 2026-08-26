package service

import (
	"context"
	"log"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/command"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/transaction"
)

// MonitorTimeout é um worker que monitora transações expiradas em background
// e aciona o rollback automático via AbortarTransacao.
type MonitorTimeout struct {
	repoTransacao transaction.Repository
	coordinador   *Coordinador2PC
	intervalo     time.Duration
}

// NovoMonitorTimeout cria uma nova instância do monitor de timeout.
func NovoMonitorTimeout(
	repoTransacao transaction.Repository,
	coordinador *Coordinador2PC,
	intervalo time.Duration,
) *MonitorTimeout {
	return &MonitorTimeout{
		repoTransacao: repoTransacao,
		coordinador:   coordinador,
		intervalo:     intervalo,
	}
}

// Iniciar dispara o loop de monitoramento em uma goroutine dedicada.
// O loop é encerrado quando o contexto for cancelado.
func (m *MonitorTimeout) Iniciar(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(m.intervalo)
		defer ticker.Stop()

		log.Printf("monitor de timeout iniciado — intervalo: %s", m.intervalo)

		for {
			select {
			case <-ctx.Done():
				log.Println("monitor de timeout encerrado")
				return
			case <-ticker.C:
				m.verificarExpiradas(ctx)
			}
		}
	}()
}

// verificarExpiradas busca transações expiradas e aciona o rollback automático.
func (m *MonitorTimeout) verificarExpiradas(ctx context.Context) {
	expiradas, err := m.repoTransacao.ListarExpiradas(ctx)
	if err != nil {
		log.Printf("erro ao listar transações expiradas: %v", err)
		return
	}

	for _, t := range expiradas {
		log.Printf("transação expirada detectada: %s — iniciando rollback automático", t.ID())

		_, err := m.coordinador.AbortarTransacao(ctx, command.AbortarTransacao{
			IDTransacao: t.ID(),
			Motivo:      "transação expirada por timeout",
		})
		if err != nil {
			log.Printf("erro ao abortar transação expirada %s: %v", t.ID(), err)
		}
	}
}
