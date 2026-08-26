package saga

import "time"

// Step representa uma etapa individual dentro de uma Saga Orquestrada.
// Cada step possui uma ação principal e uma compensação para rollback.
type Step struct {
	id             string
	idSaga         string
	nome           string
	ordem          int
	endpoint       string
	endpointCompen string
	status         StatusStep
	tentativas     int
	ultimoErro     string
	iniciadoEm     *time.Time
	concluidoEm    *time.Time
}

// NovoStep cria um novo step de Saga.
func NovoStep(idSaga, id, nome, endpoint, endpointCompen string, ordem int) *Step {
	return &Step{
		id:             id,
		idSaga:         idSaga,
		nome:           nome,
		ordem:          ordem,
		endpoint:       endpoint,
		endpointCompen: endpointCompen,
		status:         StatusStepPendente,
		tentativas:     0,
	}
}

// ReconstituirStep recria um step a partir do estado persistido.
func ReconstituirStep(
	id, idSaga, nome, endpoint, endpointCompen string,
	ordem int,
	status StatusStep,
	tentativas int,
	ultimoErro string,
	iniciadoEm *time.Time,
	concluidoEm *time.Time,
) *Step {
	return &Step{
		id:             id,
		idSaga:         idSaga,
		nome:           nome,
		ordem:          ordem,
		endpoint:       endpoint,
		endpointCompen: endpointCompen,
		status:         status,
		tentativas:     tentativas,
		ultimoErro:     ultimoErro,
		iniciadoEm:     iniciadoEm,
		concluidoEm:    concluidoEm,
	}
}

// IniciarExecucao marca o step como em execução.
func (s *Step) IniciarExecucao() {
	agora := time.Now().UTC()
	s.status = StatusStepEmExecucao
	s.iniciadoEm = &agora
	s.tentativas++
}

// MarcarComoConcluido marca o step como concluído com sucesso.
func (s *Step) MarcarComoConcluido() {
	agora := time.Now().UTC()
	s.status = StatusStepConcluido
	s.concluidoEm = &agora
}

// MarcarComoFalhou registra a falha do step com a mensagem de erro.
func (s *Step) MarcarComoFalhou(erro string) {
	s.status = StatusStepFalhou
	s.ultimoErro = erro
}

// MarcarComoCompensado marca o step como compensado após rollback.
func (s *Step) MarcarComoCompensado() {
	agora := time.Now().UTC()
	s.status = StatusStepCompensado
	s.concluidoEm = &agora
}

// Getters
func (s *Step) ID() string              { return s.id }
func (s *Step) IDSaga() string          { return s.idSaga }
func (s *Step) Nome() string            { return s.nome }
func (s *Step) Ordem() int              { return s.ordem }
func (s *Step) Endpoint() string        { return s.endpoint }
func (s *Step) EndpointCompen() string  { return s.endpointCompen }
func (s *Step) Status() StatusStep      { return s.status }
func (s *Step) Tentativas() int         { return s.tentativas }
func (s *Step) UltimoErro() string      { return s.ultimoErro }
func (s *Step) IniciadoEm() *time.Time  { return s.iniciadoEm }
func (s *Step) ConcluidoEm() *time.Time { return s.concluidoEm }
