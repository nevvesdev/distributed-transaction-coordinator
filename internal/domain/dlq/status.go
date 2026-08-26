package dlq

// Status representa o estado de uma mensagem na Dead Letter Queue.
type Status string

const (
	// StatusPendente indica que a mensagem aguarda reprocessamento.
	StatusPendente Status = "PENDENTE"
	// StatusProcessando indica que a mensagem está sendo reprocessada.
	StatusProcessando Status = "PROCESSANDO"
	// StatusResolvida indica que a mensagem foi reprocessada com sucesso.
	StatusResolvida Status = "RESOLVIDA"
	// StatusDescartada indica que a mensagem atingiu o limite de tentativas.
	StatusDescartada Status = "DESCARTADA"
)

// Terminal indica se o status é um estado final.
func (s Status) Terminal() bool {
	return s == StatusResolvida || s == StatusDescartada
}
