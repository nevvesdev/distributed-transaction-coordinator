package saga

// Status representa o estado atual de uma Saga.
type Status string

const (
	// StatusPendente indica que a Saga foi criada mas ainda não iniciou.
	StatusPendente Status = "PENDENTE"
	// StatusEmExecucao indica que a Saga está executando seus steps.
	StatusEmExecucao Status = "EM_EXECUCAO"
	// StatusConcluida indica que todos os steps foram executados com sucesso.
	StatusConcluida Status = "CONCLUIDA"
	// StatusCompensando indica que a Saga está executando as compensações.
	StatusCompensando Status = "COMPENSANDO"
	// StatusCompensada indica que todas as compensações foram executadas.
	StatusCompensada Status = "COMPENSADA"
	// StatusFalhou indica que a Saga falhou e não pôde ser compensada.
	StatusFalhou Status = "FALHOU"
)

// Terminal indica se o status é um estado final.
func (s Status) Terminal() bool {
	return s == StatusConcluida || s == StatusCompensada || s == StatusFalhou
}

// StatusStep representa o estado de um step individual da Saga.
type StatusStep string

const (
	StatusStepPendente   StatusStep = "PENDENTE"
	StatusStepEmExecucao StatusStep = "EM_EXECUCAO"
	StatusStepConcluido  StatusStep = "CONCLUIDO"
	StatusStepFalhou     StatusStep = "FALHOU"
	StatusStepCompensado StatusStep = "COMPENSADO"
)
