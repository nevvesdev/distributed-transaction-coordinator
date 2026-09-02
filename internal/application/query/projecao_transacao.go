package query

import (
	"context"
	"fmt"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/participant"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/transaction"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/persistence/eventstore"
)

// HandlerConsulta centraliza as queries de leitura da aplicação.
// Segue o princípio CQRS — leitura separada da escrita.
type HandlerConsulta struct {
	repoTransacao    transaction.Repository
	repoParticipante participant.Repository
	eventStore       eventstore.EventStore
}

// NovoHandlerConsulta cria uma nova instância do handler de consultas.
func NovoHandlerConsulta(
	repoTransacao transaction.Repository,
	repoParticipante participant.Repository,
	eventStore eventstore.EventStore,
) *HandlerConsulta {
	return &HandlerConsulta{
		repoTransacao:    repoTransacao,
		repoParticipante: repoParticipante,
		eventStore:       eventStore,
	}
}

// ConsultarAuditTrail retorna o histórico completo de eventos de um agregado.
func (h *HandlerConsulta) ConsultarAuditTrail(ctx context.Context, consulta ConsultarAuditTrail) (*ResultadoAuditTrail, error) {
	if consulta.IDAgregado == "" {
		return nil, fmt.Errorf("id do agregado é obrigatório")
	}

	registros, err := h.eventStore.ListarPorAgregado(ctx, consulta.IDAgregado)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar eventos do agregado: %w", err)
	}

	eventos := make([]EventoAudit, 0, len(registros))
	for _, r := range registros {
		eventos = append(eventos, EventoAudit{
			Sequencia:  r.ID,
			IDAgregado: r.IDAgregado,
			Evento:     r.NomeEvento,
			Payload:    string(r.Payload),
			OcorridoEm: r.OcorridoEm,
		})
	}

	return &ResultadoAuditTrail{
		IDAgregado:   consulta.IDAgregado,
		TotalEventos: len(eventos),
		Eventos:      eventos,
	}, nil
}

// ConsultarTransacao retorna o estado atual de uma transação com seus participantes.
func (h *HandlerConsulta) ConsultarTransacao(ctx context.Context, consulta ConsultarTransacao) (*ResultadoTransacao, error) {
	t, err := h.repoTransacao.BuscarPorID(ctx, consulta.ID)
	if err != nil {
		return nil, err
	}

	resultado := &ResultadoTransacao{
		ID:              t.ID(),
		Status:          string(t.Status()),
		Payload:         t.Payload(),
		ChaveIdem:       t.ChaveIdem(),
		TimeoutSegundos: int64(t.Timeout().Seconds()),
		CriadoEm:        TimestampParaString(t.CriadoEm()),
		AtualizadoEm:    TimestampParaString(t.AtualizadoEm()),
	}

	if t.ExpiradoEm() != nil {
		s := TimestampParaString(*t.ExpiradoEm())
		resultado.ExpiradoEm = &s
	}

	return resultado, nil
}

// ConsultarParticipantes retorna todos os participantes de uma transação formatados.
func (h *HandlerConsulta) ConsultarParticipantes(ctx context.Context, idTransacao string) ([]ResultadoParticipante, error) {
	participantes, err := h.repoParticipante.ListarPorTransacao(ctx, idTransacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar participantes: %w", err)
	}

	resultado := make([]ResultadoParticipante, 0, len(participantes))
	for _, p := range participantes {
		resultado = append(resultado, ResultadoParticipante{
			ID:           p.ID(),
			IDTransacao:  p.IDTransacao(),
			Endpoint:     p.Endpoint(),
			Status:       string(p.Status()),
			Tentativas:   p.Tentativas(),
			UltimoErro:   p.UltimoErro(),
			CriadoEm:     TimestampParaString(p.CriadoEm()),
			AtualizadoEm: TimestampParaString(p.AtualizadoEm()),
		})
	}

	return resultado, nil
}

// ProjetarEstadoTransacao reconstrói o estado de uma transação a partir dos eventos
// armazenados no event store — demonstra o padrão Event Sourcing puro.
func (h *HandlerConsulta) ProjetarEstadoTransacao(ctx context.Context, idTransacao string) (*ProjecaoTransacao, error) {
	registros, err := h.eventStore.ListarPorAgregado(ctx, idTransacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar eventos para projeção: %w", err)
	}

	if len(registros) == 0 {
		return nil, fmt.Errorf("nenhum evento encontrado para a transação %s", idTransacao)
	}

	projecao := &ProjecaoTransacao{
		IDTransacao: idTransacao,
		Transicoes:  make([]TransicaoEstado, 0),
	}

	for _, r := range registros {
		transicao := TransicaoEstado{
			Evento:     r.NomeEvento,
			OcorridoEm: r.OcorridoEm,
		}
		projecao.Transicoes = append(projecao.Transicoes, transicao)
		projecao.UltimoEvento = r.NomeEvento
		projecao.UltimaAtualizacao = r.OcorridoEm
	}

	projecao.TotalEventos = len(registros)
	return projecao, nil
}

// ProjecaoTransacao representa o estado reconstruído de uma transação via Event Sourcing.
type ProjecaoTransacao struct {
	IDTransacao       string            `json:"id_transacao"`
	UltimoEvento      string            `json:"ultimo_evento"`
	UltimaAtualizacao string            `json:"ultima_atualizacao"`
	TotalEventos      int               `json:"total_eventos"`
	Transicoes        []TransicaoEstado `json:"transicoes"`
}

// TransicaoEstado representa uma mudança de estado capturada como evento.
type TransicaoEstado struct {
	Evento     string `json:"evento"`
	OcorridoEm string `json:"ocorrido_em"`
}

// timeptrParaString converte *time.Time para *string formatada.
func timeptrParaString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(time.RFC3339Nano)
	return &s
}
