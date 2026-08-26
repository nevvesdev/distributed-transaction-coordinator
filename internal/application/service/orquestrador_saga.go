package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/saga"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/persistence/eventstore"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/shared/lock"
)

// DefinicaoStep contém os dados necessários para configurar um step da Saga.
type DefinicaoStep struct {
	Nome           string
	Ordem          int
	Endpoint       string
	EndpointCompen string
}

// ComandoIniciarSaga representa a entrada para criar e executar uma nova Saga.
type ComandoIniciarSaga struct {
	IDTransacao string
	NomeSaga    string
	Steps       []DefinicaoStep
}

// ResultadoSaga representa o resultado da execução de uma Saga.
type ResultadoSaga struct {
	IDSaga      string
	IDTransacao string
	Status      string
	StepAtual   int
	TotalSteps  int
}

// OrchestradorSaga implementa o orquestrador central da Saga Orquestrada.
// Executa os steps sequencialmente e aciona compensações em caso de falha.
type OrchestradorSaga struct {
	repoSaga   saga.Repository
	eventStore eventstore.EventStore
	lock       lock.DistributedLock
}

// NovoOrchestradorSaga cria uma nova instância do orquestrador.
func NovoOrchestradorSaga(
	repoSaga saga.Repository,
	eventStore eventstore.EventStore,
	lock lock.DistributedLock,
) *OrchestradorSaga {
	return &OrchestradorSaga{
		repoSaga:   repoSaga,
		eventStore: eventStore,
		lock:       lock,
	}
}

// Executar cria e executa uma Saga Orquestrada completa.
// Em caso de falha em qualquer step, aciona as compensações na ordem inversa.
func (o *OrchestradorSaga) Executar(ctx context.Context, cmd ComandoIniciarSaga) (*ResultadoSaga, error) {
	chaveLock := "saga:" + cmd.IDTransacao
	if err := o.lock.Adquirir(ctx, chaveLock); err != nil {
		return nil, fmt.Errorf("erro ao adquirir lock para saga: %w", err)
	}
	defer o.lock.Liberar(ctx, chaveLock)

	steps := make([]*saga.Step, 0, len(cmd.Steps))
	for _, def := range cmd.Steps {
		steps = append(steps, saga.NovoStep(
			"", // idSaga preenchido após criação
			uuid.NewString(),
			def.Nome,
			def.Endpoint,
			def.EndpointCompen,
			def.Ordem,
		))
	}

	s, err := saga.Nova(cmd.IDTransacao, cmd.NomeSaga, steps)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar saga: %w", err)
	}

	if err := o.repoSaga.Salvar(ctx, s); err != nil {
		return nil, fmt.Errorf("erro ao persistir saga: %w", err)
	}

	if err := s.Iniciar(); err != nil {
		return nil, fmt.Errorf("erro ao iniciar saga: %w", err)
	}

	// executa os steps sequencialmente
	for {
		step := s.ProximoStep()
		if step == nil {
			break
		}

		log.Printf("saga %s — executando step %d: %s", s.ID(), step.Ordem(), step.Nome())

		step.IniciarExecucao()
		sucesso := o.executarStep(ctx, step)

		if sucesso {
			step.MarcarComoConcluido()
			s.AvancarStep()
			if err := o.repoSaga.Atualizar(ctx, s); err != nil {
				log.Printf("aviso: erro ao atualizar saga após step %s: %v", step.Nome(), err)
			}
			continue
		}

		// falha detectada — inicia compensação
		step.MarcarComoFalhou(fmt.Sprintf("falha na execução do step %s", step.Nome()))
		log.Printf("saga %s — falha no step %d: %s — iniciando compensação", s.ID(), step.Ordem(), step.Nome())

		if err := s.IniciarCompensacao(); err != nil {
			return nil, fmt.Errorf("erro ao iniciar compensação: %w", err)
		}

		o.compensar(ctx, s)

		if err := o.repoSaga.Atualizar(ctx, s); err != nil {
			log.Printf("aviso: erro ao atualizar saga após compensação: %v", err)
		}

		return &ResultadoSaga{
			IDSaga:      s.ID(),
			IDTransacao: s.IDTransacao(),
			Status:      string(s.Status()),
			StepAtual:   s.StepAtual(),
			TotalSteps:  len(s.Steps()),
		}, nil
	}

	if err := o.repoSaga.Atualizar(ctx, s); err != nil {
		log.Printf("aviso: erro ao atualizar saga concluída: %v", err)
	}

	log.Printf("saga %s concluída com sucesso — %d steps executados", s.ID(), len(s.Steps()))

	return &ResultadoSaga{
		IDSaga:      s.ID(),
		IDTransacao: s.IDTransacao(),
		Status:      string(s.Status()),
		StepAtual:   s.StepAtual(),
		TotalSteps:  len(s.Steps()),
	}, nil
}

// compensar executa as ações de compensação na ordem inversa dos steps concluídos.
func (o *OrchestradorSaga) compensar(ctx context.Context, s *saga.Saga) {
	executados := s.StepsExecutados()

	// compensa na ordem inversa
	for i := len(executados) - 1; i >= 0; i-- {
		step := executados[i]
		log.Printf("saga %s — compensando step %d: %s", s.ID(), step.Ordem(), step.Nome())

		sucesso := o.executarCompensacao(ctx, step)
		if sucesso {
			step.MarcarComoCompensado()
		} else {
			step.MarcarComoFalhou(fmt.Sprintf("falha na compensação do step %s", step.Nome()))
			s.MarcarComoFalhou()
			log.Printf("saga %s — falha irrecuperável na compensação do step %s", s.ID(), step.Nome())
			return
		}
	}

	s.MarcarComoCompensada()
	log.Printf("saga %s — compensação concluída com sucesso", s.ID())
}

// executarStep simula o envio da ação principal ao endpoint do step.
// Na Fase 7 (HTTP Layer) será substituído por chamadas HTTP reais.
func (o *OrchestradorSaga) executarStep(ctx context.Context, step *saga.Step) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(50 * time.Millisecond):
		log.Printf("step '%s' executado no endpoint %s", step.Nome(), step.Endpoint())
		return true
	}
}

// executarCompensacao simula o envio da ação de compensação ao endpoint do step.
func (o *OrchestradorSaga) executarCompensacao(ctx context.Context, step *saga.Step) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(50 * time.Millisecond):
		log.Printf("compensação '%s' executada no endpoint %s", step.Nome(), step.EndpointCompen())
		return true
	}
}
