package query

import "time"

// ConsultarAuditTrail representa a consulta do histórico de eventos de um agregado.
type ConsultarAuditTrail struct {
	// IDAgregado é o identificador da transação ou saga a ser consultada.
	IDAgregado string
}

// EventoAudit representa um evento formatado para exibição no audit trail.
type EventoAudit struct {
	Sequencia  int64  `json:"sequencia"`
	IDAgregado string `json:"id_agregado"`
	Evento     string `json:"evento"`
	Payload    string `json:"payload"`
	OcorridoEm string `json:"ocorrido_em"`
}

// ResultadoAuditTrail é a resposta da consulta de audit trail.
type ResultadoAuditTrail struct {
	IDAgregado   string        `json:"id_agregado"`
	TotalEventos int           `json:"total_eventos"`
	Eventos      []EventoAudit `json:"eventos"`
}

// ConsultarTransacao representa a consulta do estado atual de uma transação.
type ConsultarTransacao struct {
	ID string
}

// ResultadoTransacao é a resposta com o estado atual de uma transação.
type ResultadoTransacao struct {
	ID              string            `json:"id"`
	Status          string            `json:"status"`
	Payload         map[string]string `json:"payload"`
	ChaveIdem       string            `json:"chave_idempotencia"`
	TimeoutSegundos int64             `json:"timeout_segundos"`
	CriadoEm        string            `json:"criado_em"`
	AtualizadoEm    string            `json:"atualizado_em"`
	ExpiradoEm      *string           `json:"expirado_em,omitempty"`
}

// ResultadoParticipante representa o estado de um participante na resposta HTTP.
type ResultadoParticipante struct {
	ID           string `json:"id"`
	IDTransacao  string `json:"id_transacao"`
	Endpoint     string `json:"endpoint"`
	Status       string `json:"status"`
	Tentativas   int    `json:"tentativas"`
	UltimoErro   string `json:"ultimo_erro,omitempty"`
	CriadoEm     string `json:"criado_em"`
	AtualizadoEm string `json:"atualizado_em"`
}

// TimestampParaString formata um time.Time para string RFC3339.
func TimestampParaString(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}
