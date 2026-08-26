package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/saga"
)

// RepositorioSaga implementa saga.Repository usando MySQL.
type RepositorioSaga struct {
	db *sql.DB
}

// NovoRepositorioSaga cria uma nova instância do repositório.
func NovoRepositorioSaga(db *sql.DB) *RepositorioSaga {
	return &RepositorioSaga{db: db}
}

// Salvar persiste uma nova Saga e seus steps em uma única transação.
func (r *RepositorioSaga) Salvar(ctx context.Context, s *saga.Saga) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para salvar saga: %w", err)
	}
	defer tx.Rollback()

	querySaga := `
		INSERT INTO sagas (id, id_transacao, nome, status, step_atual, criado_em, atualizado_em)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err = tx.ExecContext(ctx, querySaga,
		s.ID(),
		s.IDTransacao(),
		s.Nome(),
		string(s.Status()),
		s.StepAtual(),
		formatarTempo(s.CriadoEm()),
		formatarTempo(s.AtualizadoEm()),
	)
	if err != nil {
		return fmt.Errorf("erro ao salvar saga: %w", err)
	}

	queryStep := `
		INSERT INTO saga_steps
			(id, id_saga, nome, ordem, endpoint, endpoint_compen, status, tentativas, ultimo_erro, iniciado_em, concluido_em)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	for _, step := range s.Steps() {
		id := step.ID()
		if id == "" {
			id = uuid.NewString()
		}
		_, err = tx.ExecContext(ctx, queryStep,
			id,
			s.ID(),
			step.Nome(),
			step.Ordem(),
			step.Endpoint(),
			step.EndpointCompen(),
			string(step.Status()),
			step.Tentativas(),
			sql.NullString{String: step.UltimoErro(), Valid: step.UltimoErro() != ""},
			tempoNullable(step.IniciadoEm()),
			tempoNullable(step.ConcluidoEm()),
		)
		if err != nil {
			return fmt.Errorf("erro ao salvar step '%s': %w", step.Nome(), err)
		}
	}

	return tx.Commit()
}

// Atualizar persiste as mudanças de status da Saga e de seus steps.
func (r *RepositorioSaga) Atualizar(ctx context.Context, s *saga.Saga) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para atualizar saga: %w", err)
	}
	defer tx.Rollback()

	querySaga := `
		UPDATE sagas SET status = ?, step_atual = ?, atualizado_em = ?
		WHERE id = ?
	`
	_, err = tx.ExecContext(ctx, querySaga,
		string(s.Status()),
		s.StepAtual(),
		formatarTempo(s.AtualizadoEm()),
		s.ID(),
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar saga: %w", err)
	}

	queryStep := `
		UPDATE saga_steps
		SET status = ?, tentativas = ?, ultimo_erro = ?, iniciado_em = ?, concluido_em = ?
		WHERE id = ?
	`
	for _, step := range s.Steps() {
		_, err = tx.ExecContext(ctx, queryStep,
			string(step.Status()),
			step.Tentativas(),
			sql.NullString{String: step.UltimoErro(), Valid: step.UltimoErro() != ""},
			tempoNullable(step.IniciadoEm()),
			tempoNullable(step.ConcluidoEm()),
			step.ID(),
		)
		if err != nil {
			return fmt.Errorf("erro ao atualizar step '%s': %w", step.Nome(), err)
		}
	}

	return tx.Commit()
}

// BuscarPorID retorna uma Saga pelo seu identificador único.
func (r *RepositorioSaga) BuscarPorID(ctx context.Context, id string) (*saga.Saga, error) {
	return r.buscarSaga(ctx, "id", id)
}

// BuscarPorTransacao retorna a Saga associada a uma transação.
func (r *RepositorioSaga) BuscarPorTransacao(ctx context.Context, idTransacao string) (*saga.Saga, error) {
	return r.buscarSaga(ctx, "id_transacao", idTransacao)
}

// buscarSaga é um helper interno que busca a Saga por uma coluna específica.
func (r *RepositorioSaga) buscarSaga(ctx context.Context, coluna, valor string) (*saga.Saga, error) {
	query := fmt.Sprintf(`
		SELECT id, id_transacao, nome, status, step_atual, criado_em, atualizado_em
		FROM sagas WHERE %s = ?`, coluna)

	row := r.db.QueryRowContext(ctx, query, valor)

	var (
		id, idTransacao, nome, status string
		stepAtual                     int
		criadoEm, atualizadoEm        time.Time
	)

	err := row.Scan(&id, &idTransacao, &nome, &status, &stepAtual, &criadoEm, &atualizadoEm)
	if err == sql.ErrNoRows {
		return nil, saga.ErrSagaNaoEncontrada
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao escanear saga: %w", err)
	}

	steps, err := r.listarSteps(ctx, id)
	if err != nil {
		return nil, err
	}

	return saga.Reconstituir(
		id, idTransacao, nome,
		saga.Status(status),
		steps,
		stepAtual,
		criadoEm.UTC(),
		atualizadoEm.UTC(),
	), nil
}

// listarSteps retorna todos os steps de uma Saga ordenados pela coluna ordem.
func (r *RepositorioSaga) listarSteps(ctx context.Context, idSaga string) ([]*saga.Step, error) {
	query := `
		SELECT id, id_saga, nome, ordem, endpoint, endpoint_compen,
		       status, tentativas, ultimo_erro, iniciado_em, concluido_em
		FROM saga_steps
		WHERE id_saga = ?
		ORDER BY ordem ASC
	`

	rows, err := r.db.QueryContext(ctx, query, idSaga)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar steps da saga: %w", err)
	}
	defer rows.Close()

	var steps []*saga.Step
	for rows.Next() {
		var (
			id, idSagaCol, nome, endpoint, endpointCompen, status string
			ordem, tentativas                                     int
			ultimoErro                                            sql.NullString
			iniciadoEm, concluidoEm                               sql.NullTime
		)

		if err := rows.Scan(&id, &idSagaCol, &nome, &ordem, &endpoint, &endpointCompen,
			&status, &tentativas, &ultimoErro, &iniciadoEm, &concluidoEm); err != nil {
			return nil, fmt.Errorf("erro ao escanear step: %w", err)
		}

		var iniciadoEmPtr, concluidoEmPtr *time.Time
		if iniciadoEm.Valid {
			t := iniciadoEm.Time.UTC()
			iniciadoEmPtr = &t
		}
		if concluidoEm.Valid {
			t := concluidoEm.Time.UTC()
			concluidoEmPtr = &t
		}

		erro := ""
		if ultimoErro.Valid {
			erro = ultimoErro.String
		}

		steps = append(steps, saga.ReconstituirStep(
			id, idSagaCol, nome, endpoint, endpointCompen,
			ordem,
			saga.StatusStep(status),
			tentativas,
			erro,
			iniciadoEmPtr,
			concluidoEmPtr,
		))
	}

	return steps, rows.Err()
}

// tempoNullable converte *time.Time para interface compatível com MySQL NULL.
func tempoNullable(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return formatarTempo(*t)
}
