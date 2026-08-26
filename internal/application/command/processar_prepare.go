package command

// ProcessarPrepare representa o comando para iniciar a fase de prepare do 2PC.
type ProcessarPrepare struct {
	IDTransacao string
}

// ResultadoProcessarPrepare é a resposta do comando ProcessarPrepare.
type ResultadoProcessarPrepare struct {
	IDTransacao     string
	Status          string
	TotalPreparados int
	TotalFalharam   int
}

// ResponderPrepare representa a resposta de um participante à fase de prepare.
type ResponderPrepare struct {
	IDTransacao    string
	IDParticipante string
	// Sucesso indica se o participante confirmou o prepare.
	Sucesso bool
	// MensagemErro é preenchido quando Sucesso for false.
	MensagemErro string
}
