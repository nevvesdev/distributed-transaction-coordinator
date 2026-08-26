package command

// ProcessarCommit representa o comando para iniciar a fase de commit do 2PC.
type ProcessarCommit struct {
	IDTransacao string
}

// ResultadoProcessarCommit é a resposta do comando ProcessarCommit.
type ResultadoProcessarCommit struct {
	IDTransacao      string
	Status           string
	TotalConfirmados int
}

// AbortarTransacao representa o comando para reverter uma transação.
type AbortarTransacao struct {
	IDTransacao string
	Motivo      string
}

// ResultadoAbortarTransacao é a resposta do comando AbortarTransacao.
type ResultadoAbortarTransacao struct {
	IDTransacao string
	Status      string
	Motivo      string
}
