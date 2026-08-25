package participant

import (
	"time"

	"github.com/google/uuid"
)

// Participant é o Aggregate Root que representa um serviço participante
// em uma transação distribuída coordenada pelo protocolo 2PC ou Saga.
type Participant struct {
	id           string
	idTransacao  string
	endpoint     string
	status       Status
	tentativas   int
	ultimoErro   string
	criadoEm     time.Time
	atualizadoEm time.Time
}

// New cria um novo participante validando as regras de negócio.
func New(idTransacao, endpoint string) (*Participant, error) {
	if idTransacao == "" {
		return nil, ErrTransacaoIDInvalido
	}
	if endpoint == "" {
		return nil, ErrEndpointInvalido
	}

	agora := time.Now().UTC()
	return &Participant{
		id:           uuid.NewString(),
		idTransacao:  idTransacao,
		endpoint:     endpoint,
		status:       StatusPendente,
		tentativas:   0,
		criadoEm:     agora,
		atualizadoEm: agora,
	}, nil
}

// Reconstituir recria um participante a partir do estado persistido.
func Reconstituir(
	id string,
	idTransacao string,
	endpoint string,
	status Status,
	tentativas int,
	ultimoErro string,
	criadoEm time.Time,
	atualizadoEm time.Time,
) *Participant {
	return &Participant{
		id:           id,
		idTransacao:  idTransacao,
		endpoint:     endpoint,
		status:       status,
		tentativas:   tentativas,
		ultimoErro:   ultimoErro,
		criadoEm:     criadoEm,
		atualizadoEm: atualizadoEm,
	}
}

// MarcarComoPreparado registra que o participante confirmou a fase de prepare.
func (p *Participant) MarcarComoPreparado() error {
	if p.status != StatusPendente {
		return ErrStatusInvalido
	}
	p.status = StatusPreparado
	p.atualizadoEm = time.Now().UTC()
	return nil
}

// MarcarComoFalhou registra que o participante falhou, incrementando tentativas.
func (p *Participant) MarcarComoFalhou(erro string) {
	p.status = StatusFalhou
	p.ultimoErro = erro
	p.tentativas++
	p.atualizadoEm = time.Now().UTC()
}

// MarcarComoConfirmado registra que o participante confirmou o commit.
func (p *Participant) MarcarComoConfirmado() error {
	if p.status != StatusPreparado {
		return ErrStatusInvalido
	}
	p.status = StatusConfirmado
	p.atualizadoEm = time.Now().UTC()
	return nil
}

// MarcarComoRevertido registra que o participante executou o rollback.
func (p *Participant) MarcarComoRevertido() error {
	if p.status.Terminal() {
		return ErrStatusInvalido
	}
	p.status = StatusRevertido
	p.atualizadoEm = time.Now().UTC()
	return nil
}

// Getters
func (p *Participant) ID() string              { return p.id }
func (p *Participant) IDTransacao() string     { return p.idTransacao }
func (p *Participant) Endpoint() string        { return p.endpoint }
func (p *Participant) Status() Status          { return p.status }
func (p *Participant) Tentativas() int         { return p.tentativas }
func (p *Participant) UltimoErro() string      { return p.ultimoErro }
func (p *Participant) CriadoEm() time.Time     { return p.criadoEm }
func (p *Participant) AtualizadoEm() time.Time { return p.atualizadoEm }
