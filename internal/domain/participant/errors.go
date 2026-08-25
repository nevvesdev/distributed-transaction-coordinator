package participant

import "errors"

// Erros específicos do agregado Participant.
var (
	ErrParticipanteNaoEncontrado = errors.New("participante não encontrado")
	ErrParticipanteJaExiste      = errors.New("participante já existe")
	ErrEndpointInvalido          = errors.New("endpoint do participante é obrigatório")
	ErrTransacaoIDInvalido       = errors.New("identificador de transação é obrigatório")
	ErrStatusInvalido            = errors.New("transição de status inválida para o participante")
)
