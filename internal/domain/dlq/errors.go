package dlq

import "errors"

// Erros específicos do domínio da Dead Letter Queue.
var (
	ErrMensagemNaoEncontrada    = errors.New("mensagem não encontrada na DLQ")
	ErrMensagemJaFinalizada     = errors.New("mensagem já se encontra em estado final")
	ErrLimiteTentativasAtingido = errors.New("limite máximo de tentativas atingido")
	ErrTipoInvalido             = errors.New("tipo da mensagem é obrigatório")
	ErrPayloadInvalido          = errors.New("payload da mensagem é obrigatório")
)
