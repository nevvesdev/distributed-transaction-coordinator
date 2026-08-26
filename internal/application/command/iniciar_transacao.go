package command

import "time"

// IniciarTransacao representa o comando para criar uma nova transação distribuída.
type IniciarTransacao struct {
	// Payload contém os dados da operação a ser coordenada.
	Payload map[string]string
	// Timeout define o tempo máximo para conclusão da transação.
	Timeout time.Duration
	// ChaveIdem é a chave de idempotência para evitar duplicatas.
	ChaveIdem string
}

// ResultadoIniciarTransacao é a resposta do comando IniciarTransacao.
type ResultadoIniciarTransacao struct {
	IDTransacao string
	Status      string
	CriadoEm    string
}

// RegistrarParticipante representa o comando para adicionar um participante à transação.
type RegistrarParticipante struct {
	IDTransacao string
	Endpoint    string
}

// ResultadoRegistrarParticipante é a resposta do comando RegistrarParticipante.
type ResultadoRegistrarParticipante struct {
	IDParticipante string
	IDTransacao    string
	Endpoint       string
	Status         string
}
