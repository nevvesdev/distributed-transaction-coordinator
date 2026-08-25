package transaction

import "time"

type TransacaoIniciada struct {
	IDTransacao string
	Payload     map[string]string
	Timeout     int64
	Momento     time.Time
}

func (e TransacaoIniciada) NomeEvento() string    { return "transacao.iniciada" }
func (e TransacaoIniciada) OcorridoEm() time.Time { return e.Momento }
func (e TransacaoIniciada) IDAgregado() string    { return e.IDTransacao }

type ParticipanteRegistrado struct {
	IDTransacao    string
	IDParticipante string
	Endpoint       string
	Momento        time.Time
}

func (e ParticipanteRegistrado) NomeEvento() string    { return "transacao.participante_registrado" }
func (e ParticipanteRegistrado) OcorridoEm() time.Time { return e.Momento }
func (e ParticipanteRegistrado) IDAgregado() string    { return e.IDTransacao }

type PrepareIniciado struct {
	IDTransacao string
	Momento     time.Time
}

func (e PrepareIniciado) NomeEvento() string    { return "transacao.prepare_iniciado" }
func (e PrepareIniciado) OcorridoEm() time.Time { return e.Momento }
func (e PrepareIniciado) IDAgregado() string    { return e.IDTransacao }

type TransacaoPreparada struct {
	IDTransacao string
	Momento     time.Time
}

func (e TransacaoPreparada) NomeEvento() string    { return "transacao.preparada" }
func (e TransacaoPreparada) OcorridoEm() time.Time { return e.Momento }
func (e TransacaoPreparada) IDAgregado() string    { return e.IDTransacao }

type CommitIniciado struct {
	IDTransacao string
	Momento     time.Time
}

func (e CommitIniciado) NomeEvento() string    { return "transacao.commit_iniciado" }
func (e CommitIniciado) OcorridoEm() time.Time { return e.Momento }
func (e CommitIniciado) IDAgregado() string    { return e.IDTransacao }

type TransacaoConfirmada struct {
	IDTransacao string
	Momento     time.Time
}

func (e TransacaoConfirmada) NomeEvento() string    { return "transacao.confirmada" }
func (e TransacaoConfirmada) OcorridoEm() time.Time { return e.Momento }
func (e TransacaoConfirmada) IDAgregado() string    { return e.IDTransacao }

type TransacaoAbortada struct {
	IDTransacao string
	Motivo      string
	Momento     time.Time
}

func (e TransacaoAbortada) NomeEvento() string    { return "transacao.abortada" }
func (e TransacaoAbortada) OcorridoEm() time.Time { return e.Momento }
func (e TransacaoAbortada) IDAgregado() string    { return e.IDTransacao }

type TransacaoExpirada struct {
	IDTransacao string
	Momento     time.Time
}

func (e TransacaoExpirada) NomeEvento() string    { return "transacao.expirada" }
func (e TransacaoExpirada) OcorridoEm() time.Time { return e.Momento }
func (e TransacaoExpirada) IDAgregado() string    { return e.IDTransacao }
