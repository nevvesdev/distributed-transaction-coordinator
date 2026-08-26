package dlq

import (
	"time"

	"github.com/google/uuid"
)

// Mensagem é o Aggregate Root que representa uma mensagem na Dead Letter Queue.
// Armazena o contexto necessário para reprocessamento com retry exponencial.
type Mensagem struct {
	id               string
	idReferencia     string
	tipo             string
	payload          []byte
	status           Status
	tentativas       int
	maxTentativas    int
	ultimoErro       string
	proximaTentativa *time.Time
	criadoEm         time.Time
	atualizadoEm     time.Time
	resolvidoEm      *time.Time
}

// Nova cria uma nova mensagem na DLQ a partir de uma operação com falha.
func Nova(idReferencia, tipo string, payload []byte, maxTentativas int) (*Mensagem, error) {
	if tipo == "" {
		return nil, ErrTipoInvalido
	}
	if len(payload) == 0 {
		return nil, ErrPayloadInvalido
	}

	agora := time.Now().UTC()
	return &Mensagem{
		id:            uuid.NewString(),
		idReferencia:  idReferencia,
		tipo:          tipo,
		payload:       payload,
		status:        StatusPendente,
		tentativas:    0,
		maxTentativas: maxTentativas,
		criadoEm:      agora,
		atualizadoEm:  agora,
	}, nil
}

// Reconstituir recria uma mensagem a partir do estado persistido.
func Reconstituir(
	id, idReferencia, tipo string,
	payload []byte,
	status Status,
	tentativas, maxTentativas int,
	ultimoErro string,
	proximaTentativa *time.Time,
	criadoEm, atualizadoEm time.Time,
	resolvidoEm *time.Time,
) *Mensagem {
	return &Mensagem{
		id:               id,
		idReferencia:     idReferencia,
		tipo:             tipo,
		payload:          payload,
		status:           status,
		tentativas:       tentativas,
		maxTentativas:    maxTentativas,
		ultimoErro:       ultimoErro,
		proximaTentativa: proximaTentativa,
		criadoEm:         criadoEm,
		atualizadoEm:     atualizadoEm,
		resolvidoEm:      resolvidoEm,
	}
}

// IniciarTentativa marca a mensagem como em processamento e incrementa o contador.
func (m *Mensagem) IniciarTentativa() error {
	if m.status.Terminal() {
		return ErrMensagemJaFinalizada
	}
	if m.tentativas >= m.maxTentativas {
		return ErrLimiteTentativasAtingido
	}
	m.status = StatusProcessando
	m.tentativas++
	m.atualizadoEm = time.Now().UTC()
	return nil
}

// MarcarComoResolvida finaliza a mensagem com sucesso.
func (m *Mensagem) MarcarComoResolvida() error {
	if m.status.Terminal() {
		return ErrMensagemJaFinalizada
	}
	agora := time.Now().UTC()
	m.status = StatusResolvida
	m.resolvidoEm = &agora
	m.atualizadoEm = agora
	return nil
}

// MarcarComoFalhou registra a falha da tentativa atual e agenda a próxima.
// Se o limite for atingido, descarta a mensagem.
func (m *Mensagem) MarcarComoFalhou(erro string, proximaTentativa time.Time) {
	m.ultimoErro = erro
	m.atualizadoEm = time.Now().UTC()

	if m.tentativas >= m.maxTentativas {
		m.status = StatusDescartada
		return
	}

	m.status = StatusPendente
	m.proximaTentativa = &proximaTentativa
}

// ProntoParaReprocessar verifica se a mensagem pode ser reprocessada agora.
func (m *Mensagem) ProntoParaReprocessar() bool {
	if m.status != StatusPendente {
		return false
	}
	if m.proximaTentativa == nil {
		return true
	}
	return time.Now().UTC().After(*m.proximaTentativa)
}

// Getters
func (m *Mensagem) ID() string                   { return m.id }
func (m *Mensagem) IDReferencia() string         { return m.idReferencia }
func (m *Mensagem) Tipo() string                 { return m.tipo }
func (m *Mensagem) Payload() []byte              { return m.payload }
func (m *Mensagem) Status() Status               { return m.status }
func (m *Mensagem) Tentativas() int              { return m.tentativas }
func (m *Mensagem) MaxTentativas() int           { return m.maxTentativas }
func (m *Mensagem) UltimoErro() string           { return m.ultimoErro }
func (m *Mensagem) ProximaTentativa() *time.Time { return m.proximaTentativa }
func (m *Mensagem) CriadoEm() time.Time          { return m.criadoEm }
func (m *Mensagem) AtualizadoEm() time.Time      { return m.atualizadoEm }
func (m *Mensagem) ResolvidoEm() *time.Time      { return m.resolvidoEm }
