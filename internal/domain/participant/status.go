package participant

// Status representa o estado de um participante dentro de uma transação distribuída.
type Status string

const (
	// StatusPendente indica que o participante ainda não foi contactado.
	StatusPendente Status = "PENDENTE"
	// StatusPreparado indica que o participante confirmou o prepare com sucesso.
	StatusPreparado Status = "PREPARADO"
	// StatusFalhou indica que o participante falhou na fase de prepare.
	StatusFalhou Status = "FALHOU"
	// StatusConfirmado indica que o participante confirmou o commit com sucesso.
	StatusConfirmado Status = "CONFIRMADO"
	// StatusRevertido indica que o participante executou o rollback com sucesso.
	StatusRevertido Status = "REVERTIDO"
)

// Terminal indica se o status é um estado final sem mais transições possíveis.
func (s Status) Terminal() bool {
	return s == StatusConfirmado || s == StatusRevertido
}
