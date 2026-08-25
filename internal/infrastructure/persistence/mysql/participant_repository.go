package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/participant"
)

// RepositorioParticipante implementa participant.Repository usando MySQL.
type RepositorioParticipante struct {
	db *sql.DB
}

// NovoRepositorioParticipante cria uma nova instância do repositório.
func NovoRepositorioParticipante(db *sql.DB) *RepositorioParticipante {
	return &RepositorioParticipante{db: db}
}

// Salvar persiste um novo participante no banco de dados.
func (r *RepositorioParticipante) Salvar(ctx context.Context, p *participant.Participant) error {
	query := `
		INSERT INTO participantes
			(id, id_transacao, endpoint, status, tentativas, ultimo_erro, criado_em, atualizado_em)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?)
	`

	var ultimoErro interface{}
	if p.UltimoErro() != "" {
		ultimoErro = p.UltimoErro()
	}

	_, err := r.db.ExecContext(ctx, query,
		p.ID(),
		p.IDTransacao(),
		p.Endpoint(),
		string(p.Status()),
		p.Tentativas(),
		ultimoErro,
		formatarTempo(p.CriadoEm()),
		formatarTempo(p.AtualizadoEm()),
	)
	if err != nil {
		return fmt.Errorf("erro ao salvar participante: %w", err)
	}

	return nil
}

// Atualizar persiste as mudanças de estado de um participante existente.
func (r *RepositorioParticipante) Atualizar(ctx context.Context, p *participant.Participant) error {
	query := `
		UPDATE participantes
		SET status = ?, tentativas = ?, ultimo_erro = ?, atualizado_em = ?
		WHERE id = ?
	`

	var ultimoErro interface{}
	if p.UltimoErro() != "" {
		ultimoErro = p.UltimoErro()
	}

	resultado, err := r.db.ExecContext(ctx, query,
		string(p.Status()),
		p.Tentativas(),
		ultimoErro,
		formatarTempo(p.AtualizadoEm()),
		p.ID(),
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar participante: %w", err)
	}

	linhas, err := resultado.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}
	if linhas == 0 {
		return participant.ErrParticipanteNaoEncontrado
	}

	return nil
}

// BuscarPorID retorna um participante pelo seu identificador único.
func (r *RepositorioParticipante) BuscarPorID(ctx context.Context, id string) (*participant.Participant, error) {
	query := `
		SELECT id, id_transacao, endpoint, status, tentativas, ultimo_erro, criado_em, atualizado_em
		FROM participantes
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.escanearParticipante(row)
}

// ListarPorTransacao retorna todos os participantes de uma transação.
func (r *RepositorioParticipante) ListarPorTransacao(ctx context.Context, idTransacao string) ([]*participant.Participant, error) {
	query := `
		SELECT id, id_transacao, endpoint, status, tentativas, ultimo_erro, criado_em, atualizado_em
		FROM participantes
		WHERE id_transacao = ?
		ORDER BY criado_em ASC
	`

	rows, err := r.db.QueryContext(ctx, query, idTransacao)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar participantes: %w", err)
	}
	defer rows.Close()

	var participantes []*participant.Participant
	for rows.Next() {
		p, err := r.escanearParticipanteRows(rows)
		if err != nil {
			return nil, err
		}
		participantes = append(participantes, p)
	}

	return participantes, rows.Err()
}

// DeletarPorTransacao remove todos os participantes de uma transação.
func (r *RepositorioParticipante) DeletarPorTransacao(ctx context.Context, idTransacao string) error {
	query := `DELETE FROM participantes WHERE id_transacao = ?`

	_, err := r.db.ExecContext(ctx, query, idTransacao)
	if err != nil {
		return fmt.Errorf("erro ao deletar participantes da transação: %w", err)
	}

	return nil
}

// escanearParticipante lê um participante a partir de um *sql.Row.
func (r *RepositorioParticipante) escanearParticipante(row *sql.Row) (*participant.Participant, error) {
	var (
		id           string
		idTransacao  string
		endpoint     string
		status       string
		tentativas   int
		ultimoErro   sql.NullString
		criadoEm     time.Time
		atualizadoEm time.Time
	)

	err := row.Scan(&id, &idTransacao, &endpoint, &status, &tentativas,
		&ultimoErro, &criadoEm, &atualizadoEm)
	if err == sql.ErrNoRows {
		return nil, participant.ErrParticipanteNaoEncontrado
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao escanear participante: %w", err)
	}

	return r.reconstruirParticipante(id, idTransacao, endpoint, status,
		tentativas, ultimoErro, criadoEm, atualizadoEm), nil
}

// escanearParticipanteRows lê um participante a partir de *sql.Rows.
func (r *RepositorioParticipante) escanearParticipanteRows(rows *sql.Rows) (*participant.Participant, error) {
	var (
		id           string
		idTransacao  string
		endpoint     string
		status       string
		tentativas   int
		ultimoErro   sql.NullString
		criadoEm     time.Time
		atualizadoEm time.Time
	)

	err := rows.Scan(&id, &idTransacao, &endpoint, &status, &tentativas,
		&ultimoErro, &criadoEm, &atualizadoEm)
	if err != nil {
		return nil, fmt.Errorf("erro ao escanear participante: %w", err)
	}

	return r.reconstruirParticipante(id, idTransacao, endpoint, status,
		tentativas, ultimoErro, criadoEm, atualizadoEm), nil
}

// reconstruirParticipante monta o agregado Participant a partir dos dados lidos do banco.
func (r *RepositorioParticipante) reconstruirParticipante(
	id, idTransacao, endpoint, status string,
	tentativas int,
	ultimoErro sql.NullString,
	criadoEm, atualizadoEm time.Time,
) *participant.Participant {
	erro := ""
	if ultimoErro.Valid {
		erro = ultimoErro.String
	}

	return participant.Reconstituir(
		id,
		idTransacao,
		endpoint,
		participant.Status(status),
		tentativas,
		erro,
		criadoEm.UTC(),
		atualizadoEm.UTC(),
	)
}
