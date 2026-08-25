package transaction

type Status string

const (
	StatusIniciada    Status = "INICIADA"
	StatusPreparando  Status = "PREPARANDO"
	StatusPreparada   Status = "PREPARADA"
	StatusConfirmando Status = "CONFIRMANDO"
	StatusConfirmada  Status = "CONFIRMADA"
	StatusAbortando   Status = "ABORTANDO"
	StatusAbortada    Status = "ABORTADA"
	StatusExpirada    Status = "EXPIRADA"
)

var TransicoesValidas = map[Status][]Status{
	StatusIniciada:    {StatusPreparando, StatusAbortando},
	StatusPreparando:  {StatusPreparada, StatusAbortando, StatusExpirada},
	StatusPreparada:   {StatusConfirmando, StatusAbortando},
	StatusConfirmando: {StatusConfirmada, StatusAbortando},
	StatusAbortando:   {StatusAbortada},
}

func (s Status) PodeTransicionarPara(proximo Status) bool {
	permitidos, existe := TransicoesValidas[s]
	if !existe {
		return false
	}
	for _, permitido := range permitidos {
		if permitido == proximo {
			return true
		}
	}
	return false
}

func (s Status) Terminal() bool {
	_, possuiTransicoes := TransicoesValidas[s]
	return !possuiTransicoes
}
