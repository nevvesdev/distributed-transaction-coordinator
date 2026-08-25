package transaction

import "errors"

var (
	ErrTransacaoNaoEncontrada = errors.New("transação não encontrada")
	ErrTransacaoJaExiste      = errors.New("transação já existe")
	ErrTransicaoInvalida      = errors.New("transição de estado inválida")
	ErrTransacaoExpirada      = errors.New("transação expirada")
	ErrTransacaoJaFinalizada  = errors.New("transação já se encontra em estado final")
	ErrSemParticipantes       = errors.New("transação não possui participantes registrados")
	ErrParticipanteDuplicado  = errors.New("participante já registrado nesta transação")
	ErrTimeoutInvalido        = errors.New("timeout da transação deve ser maior que zero")
)
