package transaction

import (
	"time"

	"github.com/google/uuid"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/shared/domain"
)

// Transaction é o Aggregate Root que representa uma transação distribuída.
// Coordena o ciclo de vida completo do protocolo 2PC e da Saga Orquestrada.
type Transaction struct {
	id            string
	status        Status
	payload       map[string]string
	participantes []*Participante
	timeout       time.Duration
	chaveIdem     string
	eventos       []domain.DomainEvent
	criadoEm      time.Time
	atualizadoEm  time.Time
	expiradoEm    *time.Time
}

// Participante representa a referência a um participante dentro do agregado Transaction.
type Participante struct {
	ID       string
	Endpoint string
	Status   string
}

// New cria uma nova instância de Transaction validando as regras de negócio.
func New(payload map[string]string, timeout time.Duration, chaveIdem string) (*Transaction, error) {
	if timeout <= 0 {
		return nil, ErrTimeoutInvalido
	}

	agora := time.Now().UTC()
	expiracao := agora.Add(timeout)

	t := &Transaction{
		id:            uuid.NewString(),
		status:        StatusIniciada,
		payload:       payload,
		participantes: make([]*Participante, 0),
		timeout:       timeout,
		chaveIdem:     chaveIdem,
		eventos:       make([]domain.DomainEvent, 0),
		criadoEm:      agora,
		atualizadoEm:  agora,
		expiradoEm:    &expiracao,
	}

	t.registrarEvento(TransacaoIniciada{
		IDTransacao: t.id,
		Payload:     payload,
		Timeout:     int64(timeout.Seconds()),
		Momento:     agora,
	})

	return t, nil
}

// Reconstituir recria um agregado a partir do estado persistido (usado pelos repositórios).
func Reconstituir(
	id string,
	status Status,
	payload map[string]string,
	participantes []*Participante,
	timeout time.Duration,
	chaveIdem string,
	criadoEm time.Time,
	atualizadoEm time.Time,
	expiradoEm *time.Time,
) *Transaction {
	return &Transaction{
		id:            id,
		status:        status,
		payload:       payload,
		participantes: participantes,
		timeout:       timeout,
		chaveIdem:     chaveIdem,
		eventos:       make([]domain.DomainEvent, 0),
		criadoEm:      criadoEm,
		atualizadoEm:  atualizadoEm,
		expiradoEm:    expiradoEm,
	}
}

// AdicionarParticipante registra um novo participante na transação.
func (t *Transaction) AdicionarParticipante(id, endpoint string) error {
	if t.status.Terminal() {
		return ErrTransacaoJaFinalizada
	}
	for _, p := range t.participantes {
		if p.ID == id {
			return ErrParticipanteDuplicado
		}
	}

	t.participantes = append(t.participantes, &Participante{
		ID:       id,
		Endpoint: endpoint,
		Status:   "PENDENTE",
	})

	t.registrarEvento(ParticipanteRegistrado{
		IDTransacao:    t.id,
		IDParticipante: id,
		Endpoint:       endpoint,
		Momento:        time.Now().UTC(),
	})

	t.atualizadoEm = time.Now().UTC()
	return nil
}

func (t *Transaction) IniciarPrepare() error {
	if len(t.participantes) == 0 {
		return ErrSemParticipantes
	}
	return t.transicionarPara(StatusPreparando, PrepareIniciado{
		IDTransacao: t.id,
		Momento:     time.Now().UTC(),
	})
}

func (t *Transaction) MarcarComoPreparada() error {
	return t.transicionarPara(StatusPreparada, TransacaoPreparada{
		IDTransacao: t.id,
		Momento:     time.Now().UTC(),
	})
}

func (t *Transaction) IniciarCommit() error {
	return t.transicionarPara(StatusConfirmando, CommitIniciado{
		IDTransacao: t.id,
		Momento:     time.Now().UTC(),
	})
}

func (t *Transaction) MarcarComoConfirmada() error {
	return t.transicionarPara(StatusConfirmada, TransacaoConfirmada{
		IDTransacao: t.id,
		Momento:     time.Now().UTC(),
	})
}

func (t *Transaction) Abortar(motivo string) error {
	if t.status.Terminal() {
		return ErrTransacaoJaFinalizada
	}
	return t.transicionarPara(StatusAbortando, TransacaoAbortada{
		IDTransacao: t.id,
		Motivo:      motivo,
		Momento:     time.Now().UTC(),
	})
}

func (t *Transaction) MarcarComoAbortada() error {
	return t.transicionarPara(StatusAbortada, TransacaoAbortada{
		IDTransacao: t.id,
		Motivo:      "rollback concluído",
		Momento:     time.Now().UTC(),
	})
}

func (t *Transaction) Expirar() error {
	return t.transicionarPara(StatusExpirada, TransacaoExpirada{
		IDTransacao: t.id,
		Momento:     time.Now().UTC(),
	})
}

func (t *Transaction) EstaExpirada() bool {
	if t.expiradoEm == nil {
		return false
	}
	return time.Now().UTC().After(*t.expiradoEm)
}

func (t *Transaction) ColetarEventos() []domain.DomainEvent {
	eventos := t.eventos
	t.eventos = make([]domain.DomainEvent, 0)
	return eventos
}

func (t *Transaction) ID() string                     { return t.id }
func (t *Transaction) Status() Status                 { return t.status }
func (t *Transaction) Payload() map[string]string     { return t.payload }
func (t *Transaction) Participantes() []*Participante { return t.participantes }
func (t *Transaction) Timeout() time.Duration         { return t.timeout }
func (t *Transaction) ChaveIdem() string              { return t.chaveIdem }
func (t *Transaction) CriadoEm() time.Time            { return t.criadoEm }
func (t *Transaction) AtualizadoEm() time.Time        { return t.atualizadoEm }
func (t *Transaction) ExpiradoEm() *time.Time         { return t.expiradoEm }

func (t *Transaction) transicionarPara(novoStatus Status, evento domain.DomainEvent) error {
	if !t.status.PodeTransicionarPara(novoStatus) {
		return ErrTransicaoInvalida
	}
	t.status = novoStatus
	t.atualizadoEm = time.Now().UTC()
	t.registrarEvento(evento)
	return nil
}

func (t *Transaction) registrarEvento(evento domain.DomainEvent) {
	t.eventos = append(t.eventos, evento)
}
