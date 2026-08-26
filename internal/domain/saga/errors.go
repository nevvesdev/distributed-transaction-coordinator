package saga

import "errors"

// Erros específicos do agregado Saga.
var (
	ErrSagaNaoEncontrada = errors.New("saga não encontrada")
	ErrSagaJaFinalizada  = errors.New("saga já se encontra em estado final")
	ErrSemSteps          = errors.New("saga não possui steps definidos")
	ErrStepInvalido      = errors.New("step inválido ou não encontrado")
	ErrOrdemStepInvalida = errors.New("ordem dos steps deve ser sequencial a partir de 1")
)
