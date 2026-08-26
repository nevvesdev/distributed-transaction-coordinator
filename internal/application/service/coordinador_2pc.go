package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/application/command"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/participant"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/transaction"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/persistence/eventstore"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/shared/lock"
)

// Coordinador2PC implementa o protocolo Two-Phase Commit como application service.
// Coordena o ciclo de vida completo: prepare → commit ou abort.
type Coordinador2PC struct {
	repoTransacao    transaction.Repository
	repoParticipante participant.Repository
	eventStore       eventstore.EventStore
	lock             lock.DistributedLock
	timeoutPrepare   time.Duration
}

// NovoCoordinador2PC cria uma nova instância do coordinador.
func NovoCoordinador2PC(
	repoTransacao transaction.Repository,
	repoParticipante participant.Repository,
	eventStore eventstore.EventStore,
	lock lock.DistributedLock,
	timeoutPrepare time.Duration,
) *Coordinador2PC {
	return &Coordinador2PC{
		repoTransacao:    repoTransacao,
		repoParticipante: repoParticipante,
		eventStore:       eventStore,
		lock:             lock,
		timeoutPrepare:   timeoutPrepare,
	}
}

// IniciarTransacao processa o comando de criação de uma nova transação distribuída.
func (c *Coordinador2PC) IniciarTransacao(ctx context.Context, cmd command.IniciarTransacao) (*command.ResultadoIniciarTransacao, error) {
	// verifica idempotência antes de qualquer operação
	existente, err := c.repoTransacao.BuscarPorChaveIdem(ctx, cmd.ChaveIdem)
	if err != nil && err != transaction.ErrTransacaoNaoEncontrada {
		return nil, fmt.Errorf("erro ao verificar idempotência: %w", err)
	}
	if existente != nil {
		return &command.ResultadoIniciarTransacao{
			IDTransacao: existente.ID(),
			Status:      string(existente.Status()),
			CriadoEm:    existente.CriadoEm().Format(time.RFC3339),
		}, nil
	}

	t, err := transaction.New(cmd.Payload, cmd.Timeout, cmd.ChaveIdem)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar transação: %w", err)
	}

	if err := c.repoTransacao.Salvar(ctx, t); err != nil {
		return nil, fmt.Errorf("erro ao persistir transação: %w", err)
	}

	if err := c.eventStore.Salvar(ctx, t.ColetarEventos()); err != nil {
		log.Printf("aviso: erro ao salvar eventos da transação %s: %v", t.ID(), err)
	}

	return &command.ResultadoIniciarTransacao{
		IDTransacao: t.ID(),
		Status:      string(t.Status()),
		CriadoEm:    t.CriadoEm().Format(time.RFC3339),
	}, nil
}

// RegistrarParticipante adiciona um participante à transação informada.
func (c *Coordinador2PC) RegistrarParticipante(ctx context.Context, cmd command.RegistrarParticipante) (*command.ResultadoRegistrarParticipante, error) {
	chaveLock := "transacao:" + cmd.IDTransacao
	if err := c.lock.Adquirir(ctx, chaveLock); err != nil {
		return nil, fmt.Errorf("erro ao adquirir lock da transação: %w", err)
	}
	defer c.lock.Liberar(ctx, chaveLock)

	t, err := c.repoTransacao.BuscarPorID(ctx, cmd.IDTransacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar transação: %w", err)
	}

	p, err := participant.New(cmd.IDTransacao, cmd.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar participante: %w", err)
	}

	if err := t.AdicionarParticipante(p.ID(), p.Endpoint()); err != nil {
		return nil, fmt.Errorf("erro ao adicionar participante à transação: %w", err)
	}

	if err := c.repoParticipante.Salvar(ctx, p); err != nil {
		return nil, fmt.Errorf("erro ao persistir participante: %w", err)
	}

	if err := c.repoTransacao.Atualizar(ctx, t); err != nil {
		return nil, fmt.Errorf("erro ao atualizar transação: %w", err)
	}

	if err := c.eventStore.Salvar(ctx, t.ColetarEventos()); err != nil {
		log.Printf("aviso: erro ao salvar eventos do participante %s: %v", p.ID(), err)
	}

	return &command.ResultadoRegistrarParticipante{
		IDParticipante: p.ID(),
		IDTransacao:    p.IDTransacao(),
		Endpoint:       p.Endpoint(),
		Status:         string(p.Status()),
	}, nil
}

// ProcessarPrepare executa a fase 1 do 2PC — envia prepare a todos os participantes.
func (c *Coordinador2PC) ProcessarPrepare(ctx context.Context, cmd command.ProcessarPrepare) (*command.ResultadoProcessarPrepare, error) {
	chaveLock := "transacao:" + cmd.IDTransacao
	if err := c.lock.Adquirir(ctx, chaveLock); err != nil {
		return nil, fmt.Errorf("erro ao adquirir lock para prepare: %w", err)
	}
	defer c.lock.Liberar(ctx, chaveLock)

	t, err := c.repoTransacao.BuscarPorID(ctx, cmd.IDTransacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar transação: %w", err)
	}

	if t.EstaExpirada() {
		_ = t.Expirar()
		_ = c.repoTransacao.Atualizar(ctx, t)
		return nil, transaction.ErrTransacaoExpirada
	}

	if err := t.IniciarPrepare(); err != nil {
		return nil, fmt.Errorf("erro ao iniciar fase de prepare: %w", err)
	}

	if err := c.repoTransacao.Atualizar(ctx, t); err != nil {
		return nil, fmt.Errorf("erro ao atualizar status para PREPARANDO: %w", err)
	}

	participantes, err := c.repoParticipante.ListarPorTransacao(ctx, cmd.IDTransacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar participantes: %w", err)
	}

	preparados := 0
	falharam := 0

	ctxPrepare, cancelar := context.WithTimeout(ctx, c.timeoutPrepare)
	defer cancelar()

	for _, p := range participantes {
		sucesso := c.enviarPrepare(ctxPrepare, p.Endpoint(), cmd.IDTransacao)
		if sucesso {
			_ = p.MarcarComoPreparado()
			preparados++
		} else {
			p.MarcarComoFalhou("falha na fase de prepare")
			falharam++
		}
		_ = c.repoParticipante.Atualizar(ctx, p)
	}

	if falharam > 0 {
		motivo := fmt.Sprintf("%d participante(s) falharam na fase de prepare", falharam)
		_ = t.Abortar(motivo)
		_ = c.repoTransacao.Atualizar(ctx, t)
		if err := c.eventStore.Salvar(ctx, t.ColetarEventos()); err != nil {
			log.Printf("aviso: erro ao salvar eventos de abort: %v", err)
		}
		return &command.ResultadoProcessarPrepare{
			IDTransacao:     cmd.IDTransacao,
			Status:          string(t.Status()),
			TotalPreparados: preparados,
			TotalFalharam:   falharam,
		}, nil
	}

	_ = t.MarcarComoPreparada()
	_ = c.repoTransacao.Atualizar(ctx, t)

	if err := c.eventStore.Salvar(ctx, t.ColetarEventos()); err != nil {
		log.Printf("aviso: erro ao salvar eventos de prepare: %v", err)
	}

	return &command.ResultadoProcessarPrepare{
		IDTransacao:     cmd.IDTransacao,
		Status:          string(t.Status()),
		TotalPreparados: preparados,
		TotalFalharam:   falharam,
	}, nil
}

// ProcessarCommit executa a fase 2 do 2PC — envia commit a todos os participantes preparados.
func (c *Coordinador2PC) ProcessarCommit(ctx context.Context, cmd command.ProcessarCommit) (*command.ResultadoProcessarCommit, error) {
	chaveLock := "transacao:" + cmd.IDTransacao
	if err := c.lock.Adquirir(ctx, chaveLock); err != nil {
		return nil, fmt.Errorf("erro ao adquirir lock para commit: %w", err)
	}
	defer c.lock.Liberar(ctx, chaveLock)

	t, err := c.repoTransacao.BuscarPorID(ctx, cmd.IDTransacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar transação: %w", err)
	}

	if err := t.IniciarCommit(); err != nil {
		return nil, fmt.Errorf("erro ao iniciar fase de commit: %w", err)
	}

	if err := c.repoTransacao.Atualizar(ctx, t); err != nil {
		return nil, fmt.Errorf("erro ao atualizar status para CONFIRMANDO: %w", err)
	}

	participantes, err := c.repoParticipante.ListarPorTransacao(ctx, cmd.IDTransacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar participantes: %w", err)
	}

	confirmados := 0
	for _, p := range participantes {
		if p.Status() != participant.StatusPreparado {
			continue
		}
		sucesso := c.enviarCommit(ctx, p.Endpoint(), cmd.IDTransacao)
		if sucesso {
			_ = p.MarcarComoConfirmado()
			confirmados++
		} else {
			p.MarcarComoFalhou("falha na fase de commit")
		}
		_ = c.repoParticipante.Atualizar(ctx, p)
	}

	_ = t.MarcarComoConfirmada()
	_ = c.repoTransacao.Atualizar(ctx, t)

	if err := c.eventStore.Salvar(ctx, t.ColetarEventos()); err != nil {
		log.Printf("aviso: erro ao salvar eventos de commit: %v", err)
	}

	return &command.ResultadoProcessarCommit{
		IDTransacao:      cmd.IDTransacao,
		Status:           string(t.Status()),
		TotalConfirmados: confirmados,
	}, nil
}

// AbortarTransacao reverte a transação e notifica todos os participantes.
func (c *Coordinador2PC) AbortarTransacao(ctx context.Context, cmd command.AbortarTransacao) (*command.ResultadoAbortarTransacao, error) {
	chaveLock := "transacao:" + cmd.IDTransacao
	if err := c.lock.Adquirir(ctx, chaveLock); err != nil {
		return nil, fmt.Errorf("erro ao adquirir lock para abort: %w", err)
	}
	defer c.lock.Liberar(ctx, chaveLock)

	t, err := c.repoTransacao.BuscarPorID(ctx, cmd.IDTransacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar transação: %w", err)
	}

	if err := t.Abortar(cmd.Motivo); err != nil {
		return nil, fmt.Errorf("erro ao abortar transação: %w", err)
	}

	participantes, err := c.repoParticipante.ListarPorTransacao(ctx, cmd.IDTransacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar participantes para rollback: %w", err)
	}

	for _, p := range participantes {
		if p.Status().Terminal() {
			continue
		}
		c.enviarRollback(ctx, p.Endpoint(), cmd.IDTransacao)
		_ = p.MarcarComoRevertido()
		_ = c.repoParticipante.Atualizar(ctx, p)
	}

	_ = t.MarcarComoAbortada()
	_ = c.repoTransacao.Atualizar(ctx, t)

	if err := c.eventStore.Salvar(ctx, t.ColetarEventos()); err != nil {
		log.Printf("aviso: erro ao salvar eventos de abort: %v", err)
	}

	return &command.ResultadoAbortarTransacao{
		IDTransacao: cmd.IDTransacao,
		Status:      string(t.Status()),
		Motivo:      cmd.Motivo,
	}, nil
}

// enviarPrepare simula o envio do comando prepare ao endpoint do participante.
// Na Fase 7 (HTTP Layer) esta função será substituída por chamadas HTTP reais.
func (c *Coordinador2PC) enviarPrepare(ctx context.Context, endpoint, idTransacao string) bool {
	log.Printf("enviando prepare para %s — transação %s", endpoint, idTransacao)
	select {
	case <-ctx.Done():
		log.Printf("timeout ao enviar prepare para %s", endpoint)
		return false
	case <-time.After(50 * time.Millisecond):
		return true
	}
}

// enviarCommit simula o envio do comando commit ao endpoint do participante.
func (c *Coordinador2PC) enviarCommit(ctx context.Context, endpoint, idTransacao string) bool {
	log.Printf("enviando commit para %s — transação %s", endpoint, idTransacao)
	select {
	case <-ctx.Done():
		return false
	case <-time.After(50 * time.Millisecond):
		return true
	}
}

// enviarRollback simula o envio do comando rollback ao endpoint do participante.
func (c *Coordinador2PC) enviarRollback(ctx context.Context, endpoint, idTransacao string) {
	log.Printf("enviando rollback para %s — transação %s", endpoint, idTransacao)
	select {
	case <-ctx.Done():
	case <-time.After(50 * time.Millisecond):
	}
}
