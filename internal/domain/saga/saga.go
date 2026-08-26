package saga

import (
	"sort"
	"time"

	"github.com/google/uuid"
)

// Saga é o Aggregate Root que representa uma Saga Orquestrada.
// Gerencia o fluxo sequencial de steps e suas compensações em caso de falha.
type Saga struct {
	id           string
	idTransacao  string
	nome         string
	status       Status
	steps        []*Step
	stepAtual    int
	criadoEm     time.Time
	atualizadoEm time.Time
}

// Nova cria uma nova instância de Saga com os steps definidos.
func Nova(idTransacao, nome string, steps []*Step) (*Saga, error) {
	if len(steps) == 0 {
		return nil, ErrSemSteps
	}

	// garante ordenação sequencial
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Ordem() < steps[j].Ordem()
	})

	for i, s := range steps {
		if s.Ordem() != i+1 {
			return nil, ErrOrdemStepInvalida
		}
	}

	agora := time.Now().UTC()
	return &Saga{
		id:           uuid.NewString(),
		idTransacao:  idTransacao,
		nome:         nome,
		status:       StatusPendente,
		steps:        steps,
		stepAtual:    0,
		criadoEm:     agora,
		atualizadoEm: agora,
	}, nil
}

// Reconstituir recria uma Saga a partir do estado persistido.
func Reconstituir(
	id, idTransacao, nome string,
	status Status,
	steps []*Step,
	stepAtual int,
	criadoEm, atualizadoEm time.Time,
) *Saga {
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Ordem() < steps[j].Ordem()
	})
	return &Saga{
		id:           id,
		idTransacao:  idTransacao,
		nome:         nome,
		status:       status,
		steps:        steps,
		stepAtual:    stepAtual,
		criadoEm:     criadoEm,
		atualizadoEm: atualizadoEm,
	}
}

// Iniciar avança a Saga para o estado em execução.
func (s *Saga) Iniciar() error {
	if s.status != StatusPendente {
		return ErrSagaJaFinalizada
	}
	s.status = StatusEmExecucao
	s.atualizadoEm = time.Now().UTC()
	return nil
}

// ProximoStep retorna o próximo step a ser executado, ou nil se não houver.
func (s *Saga) ProximoStep() *Step {
	if s.stepAtual >= len(s.steps) {
		return nil
	}
	return s.steps[s.stepAtual]
}

// AvancarStep registra o sucesso do step atual e avança o índice.
func (s *Saga) AvancarStep() {
	s.stepAtual++
	s.atualizadoEm = time.Now().UTC()

	if s.stepAtual >= len(s.steps) {
		s.status = StatusConcluida
	}
}

// IniciarCompensacao avança a Saga para o estado de compensação.
func (s *Saga) IniciarCompensacao() error {
	if s.status.Terminal() {
		return ErrSagaJaFinalizada
	}
	s.status = StatusCompensando
	s.atualizadoEm = time.Now().UTC()
	return nil
}

// MarcarComoCompensada finaliza o processo de compensação com sucesso.
func (s *Saga) MarcarComoCompensada() {
	s.status = StatusCompensada
	s.atualizadoEm = time.Now().UTC()
}

// MarcarComoFalhou registra falha irrecuperável na Saga.
func (s *Saga) MarcarComoFalhou() {
	s.status = StatusFalhou
	s.atualizadoEm = time.Now().UTC()
}

// StepsExecutados retorna todos os steps executados com sucesso até o momento.
func (s *Saga) StepsExecutados() []*Step {
	executados := make([]*Step, 0)
	for _, step := range s.steps {
		if step.Status() == StatusStepConcluido {
			executados = append(executados, step)
		}
	}
	return executados
}

// Getters
func (s *Saga) ID() string              { return s.id }
func (s *Saga) IDTransacao() string     { return s.idTransacao }
func (s *Saga) Nome() string            { return s.nome }
func (s *Saga) Status() Status          { return s.status }
func (s *Saga) Steps() []*Step          { return s.steps }
func (s *Saga) StepAtual() int          { return s.stepAtual }
func (s *Saga) CriadoEm() time.Time     { return s.criadoEm }
func (s *Saga) AtualizadoEm() time.Time { return s.atualizadoEm }
