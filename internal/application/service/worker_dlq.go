package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/command"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/dlq"
)

const (
	// limiteMensagensPorCiclo define quantas mensagens são processadas por execução do worker.
	limiteMensagensPorCiclo = 10
	// fatorJitter é o percentual máximo de aleatoriedade aplicado ao backoff.
	fatorJitter = 0.3
)

// ProcessadorMensagem define a função responsável por reprocessar uma mensagem da DLQ.
type ProcessadorMensagem func(ctx context.Context, tipo string, payload []byte) error

// WorkerDLQ é o worker responsável por monitorar e reprocessar mensagens da DLQ
// usando retry com backoff exponencial e jitter para evitar thundering herd.
type WorkerDLQ struct {
	repoDLQ       dlq.Repository
	coordinador   *Coordinador2PC
	orquestrador  *OrchestradorSaga
	intervalo     time.Duration
	intervaloBase time.Duration
	maxTentativas int
}

// NovoWorkerDLQ cria uma nova instância do worker da DLQ.
func NovoWorkerDLQ(
	repoDLQ dlq.Repository,
	coordinador *Coordinador2PC,
	orquestrador *OrchestradorSaga,
	intervalo time.Duration,
	intervaloBase time.Duration,
	maxTentativas int,
) *WorkerDLQ {
	return &WorkerDLQ{
		repoDLQ:       repoDLQ,
		coordinador:   coordinador,
		orquestrador:  orquestrador,
		intervalo:     intervalo,
		intervaloBase: intervaloBase,
		maxTentativas: maxTentativas,
	}
}

// Iniciar dispara o loop de reprocessamento em uma goroutine dedicada.
func (w *WorkerDLQ) Iniciar(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(w.intervalo)
		defer ticker.Stop()

		log.Printf("worker DLQ iniciado — intervalo: %s, base retry: %s, max tentativas: %d",
			w.intervalo, w.intervaloBase, w.maxTentativas)

		for {
			select {
			case <-ctx.Done():
				log.Println("worker DLQ encerrado")
				return
			case <-ticker.C:
				w.processarCiclo(ctx)
			}
		}
	}()
}

// processarCiclo busca e reprocessa as mensagens pendentes do ciclo atual.
func (w *WorkerDLQ) processarCiclo(ctx context.Context) {
	pendentes, err := w.repoDLQ.ListarPendentes(ctx, limiteMensagensPorCiclo)
	if err != nil {
		log.Printf("worker DLQ — erro ao buscar mensagens pendentes: %v", err)
		return
	}

	if len(pendentes) == 0 {
		return
	}

	log.Printf("worker DLQ — processando %d mensagem(ns) pendente(s)", len(pendentes))

	for _, mensagem := range pendentes {
		w.reprocessar(ctx, mensagem)
	}
}

// reprocessar tenta reprocessar uma mensagem individual da DLQ.
func (w *WorkerDLQ) reprocessar(ctx context.Context, mensagem *dlq.Mensagem) {
	if err := mensagem.IniciarTentativa(); err != nil {
		log.Printf("worker DLQ — mensagem %s não pode ser reprocessada: %v", mensagem.ID(), err)
		_ = w.repoDLQ.Atualizar(ctx, mensagem)
		return
	}

	log.Printf("worker DLQ — tentativa %d/%d para mensagem %s (tipo: %s)",
		mensagem.Tentativas(), mensagem.MaxTentativas(), mensagem.ID(), mensagem.Tipo())

	err := w.despachar(ctx, mensagem)

	if err == nil {
		_ = mensagem.MarcarComoResolvida()
		log.Printf("worker DLQ — mensagem %s resolvida com sucesso", mensagem.ID())
	} else {
		proximaTentativa := w.calcularProximaTentativa(mensagem.Tentativas())
		mensagem.MarcarComoFalhou(err.Error(), proximaTentativa)
		log.Printf("worker DLQ — mensagem %s falhou (tentativa %d/%d) — próxima em: %s — erro: %v",
			mensagem.ID(), mensagem.Tentativas(), mensagem.MaxTentativas(),
			proximaTentativa.Format(time.RFC3339), err)
	}

	if err := w.repoDLQ.Atualizar(ctx, mensagem); err != nil {
		log.Printf("worker DLQ — erro ao atualizar mensagem %s: %v", mensagem.ID(), err)
	}
}

// despachar roteia a mensagem para o handler correto com base no tipo.
func (w *WorkerDLQ) despachar(ctx context.Context, mensagem *dlq.Mensagem) error {
	switch mensagem.Tipo() {
	case "2pc.commit":
		return w.reprocessarCommit(ctx, mensagem.IDReferencia(), mensagem.Payload())
	case "2pc.abort":
		return w.reprocessarAbort(ctx, mensagem.IDReferencia(), mensagem.Payload())
	case "saga.executar":
		return w.reprocessarSaga(ctx, mensagem.IDReferencia(), mensagem.Payload())
	default:
		return fmt.Errorf("tipo de mensagem desconhecido: %s", mensagem.Tipo())
	}
}

// reprocessarCommit tenta reenviar o comando de commit para a transação.
func (w *WorkerDLQ) reprocessarCommit(ctx context.Context, idTransacao string, _ []byte) error {
	_, err := w.coordinador.ProcessarCommit(ctx, command.ProcessarCommit{
		IDTransacao: idTransacao,
	})
	return err
}

// reprocessarAbort tenta reenviar o comando de abort para a transação.
func (w *WorkerDLQ) reprocessarAbort(ctx context.Context, idTransacao string, payload []byte) error {
	var dados map[string]string
	if err := json.Unmarshal(payload, &dados); err != nil {
		return fmt.Errorf("erro ao desserializar payload de abort: %w", err)
	}

	motivo := dados["motivo"]
	if motivo == "" {
		motivo = "reprocessamento via DLQ"
	}

	_, err := w.coordinador.AbortarTransacao(ctx, command.AbortarTransacao{
		IDTransacao: idTransacao,
		Motivo:      motivo,
	})
	return err
}

// reprocessarSaga tenta reexecutar uma Saga a partir do payload original.
func (w *WorkerDLQ) reprocessarSaga(ctx context.Context, idTransacao string, payload []byte) error {
	var cmd ComandoIniciarSaga
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return fmt.Errorf("erro ao desserializar payload da saga: %w", err)
	}
	cmd.IDTransacao = idTransacao

	_, err := w.orquestrador.Executar(ctx, cmd)
	return err
}

// calcularProximaTentativa aplica backoff exponencial com jitter para determinar
// o momento da próxima tentativa de reprocessamento.
//
// Fórmula: base * 2^(tentativa-1) * (1 + jitter_aleatório)
// Exemplos com base=1s:
//
//	tentativa 1 → ~1s
//	tentativa 2 → ~2s
//	tentativa 3 → ~4s
//	tentativa 4 → ~8s
//	tentativa 5 → ~16s
func (w *WorkerDLQ) calcularProximaTentativa(tentativa int) time.Time {
	expoente := math.Pow(2, float64(tentativa-1))
	backoff := float64(w.intervaloBase) * expoente

	// aplica jitter aleatório para evitar thundering herd
	jitter := backoff * fatorJitter * rand.Float64()
	duracao := time.Duration(backoff + jitter)

	return time.Now().UTC().Add(duracao)
}

// EnfileirarMensagem cria e persiste uma nova mensagem na DLQ.
// Deve ser chamado quando uma operação falha e precisa de retry.
func (w *WorkerDLQ) EnfileirarMensagem(ctx context.Context, idReferencia, tipo string, payload []byte) error {
	mensagem, err := dlq.Nova(idReferencia, tipo, payload, w.maxTentativas)
	if err != nil {
		return fmt.Errorf("erro ao criar mensagem para DLQ: %w", err)
	}

	if err := w.repoDLQ.Salvar(ctx, mensagem); err != nil {
		return fmt.Errorf("erro ao enfileirar mensagem na DLQ: %w", err)
	}

	log.Printf("worker DLQ — mensagem enfileirada: id=%s tipo=%s referencia=%s",
		mensagem.ID(), tipo, idReferencia)

	return nil
}
