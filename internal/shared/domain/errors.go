package domain

import "errors"

var (
	ErrIDInvalido            = errors.New("identificador inválido")
	ErrOperacaoInvalida      = errors.New("operação inválida para o estado atual")
	ErrEntidadeNaoEncontrada = errors.New("entidade não encontrada")
)
