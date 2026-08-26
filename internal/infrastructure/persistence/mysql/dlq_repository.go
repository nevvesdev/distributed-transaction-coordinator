package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/dlq"
)

// RepositorioDLQ implementa dlq.Repository usando MySQL.
type RepositorioDLQ struct {
	db *sql.DB
}

// NovoRepositorioDLQ cria uma nova instância do repositório da DLQ.
func NovoRepositorioDLQ(db *sql.DB) *RepositorioDLQ {
	return &RepositorioDLQ{db: db}
}

// Salvar persiste uma nova mensagem na DLQ.
func (r *RepositorioDLQ) Salvar(ctx context.Context, m *dlq.Mensagem) error {
	query := `
		INSERT INTO dlq_mensagens
			(id, id_referencia, tipo, payload, status, tentativas, max_tentativas,
			 ultimo_erro, proxima_tentativa, criado_em, atualizado_em, resolvido_em)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		m.ID(),
		m.IDReferencia(),
		m.Tipo(),
		m.Payload(),
		string(m.Status()),
		m.Tentativas(),
		m.MaxTentativas(),
		sql.NullString{String: m.UltimoErro(), Valid: m.UltimoErro() != ""},
		tempoNullable(m.ProximaTentativa()),
		formatarTempo(m.CriadoEm()),
		formatarTempo(m.AtualizadoEm()),
		tempoNullable(m.ResolvidoEm()),
	)
	if err != nil {
		return fmt.Errorf("erro ao salvar mensagem na DLQ: %w", err)
	}

	return nil
}

// Atualizar persiste as mudanças de estado de uma mensagem existente.
func (r *RepositorioDLQ) Atualizar(ctx context.Context, m *dlq.Mensagem) error {
	query := `
		UPDATE dlq_mensagens
		SET status = ?, tentativas = ?, ultimo_erro = ?,
		    proxima_tentativa = ?, atualizado_em = ?, resolvido_em = ?
		WHERE id = ?
	`

	resultado, err := r.db.ExecContext(ctx, query,
		string(m.Status()),
		m.Tentativas(),
		sql.NullString{String: m.UltimoErro(), Valid: m.UltimoErro() != ""},
		tempoNullable(m.ProximaTentativa()),
		formatarTempo(m.AtualizadoEm()),
		tempoNullable(m.ResolvidoEm()),
		m.ID(),
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar mensagem na DLQ: %w", err)
	}

	linhas, err := resultado.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}
	if linhas == 0 {
		return dlq.ErrMensagemNaoEncontrada
	}

	return nil
}

// BuscarPorID retorna uma mensagem pelo seu identificador único.
func (r *RepositorioDLQ) BuscarPorID(ctx context.Context, id string) (*dlq.Mensagem, error) {
	query := `
		SELECT id, id_referencia, tipo, payload, status, tentativas, max_tentativas,
		       ultimo_erro, proxima_tentativa, criado_em, atualizado_em, resolvido_em
		FROM dlq_mensagens
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.escanearMensagem(row)
}

// ListarPendentes retorna mensagens prontas para reprocessamento agora.
func (r *RepositorioDLQ) ListarPendentes(ctx context.Context, limite int) ([]*dlq.Mensagem, error) {
	query := `
		SELECT id, id_referencia, tipo, payload, status, tentativas, max_tentativas,
		       ultimo_erro, proxima_tentativa, criado_em, atualizado_em, resolvido_em
		FROM dlq_mensagens
		WHERE status = 'PENDENTE'
		  AND (proxima_tentativa IS NULL OR proxima_tentativa <= ?)
		ORDER BY criado_em ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, formatarTempo(time.Now().UTC()), limite)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar mensagens pendentes da DLQ: %w", err)
	}
	defer rows.Close()

	return r.escanearMensagens(rows)
}

// ListarPorReferencia retorna todas as mensagens de uma referência.
func (r *RepositorioDLQ) ListarPorReferencia(ctx context.Context, idReferencia string) ([]*dlq.Mensagem, error) {
	query := `
		SELECT id, id_referencia, tipo, payload, status, tentativas, max_tentativas,
		       ultimo_erro, proxima_tentativa, criado_em, atualizado_em, resolvido_em
		FROM dlq_mensagens
		WHERE id_referencia = ?
		ORDER BY criado_em ASC
	`

	rows, err := r.db.QueryContext(ctx, query, idReferencia)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar mensagens por referência: %w", err)
	}
	defer rows.Close()

	return r.escanearMensagens(rows)
}

// escanearMensagem lê uma mensagem a partir de um *sql.Row.
func (r *RepositorioDLQ) escanearMensagem(row *sql.Row) (*dlq.Mensagem, error) {
	var (
		id, idReferencia, tipo, status string
		payload                        []byte
		tentativas, maxTentativas      int
		ultimoErro                     sql.NullString
		proximaTentativa               sql.NullTime
		criadoEm, atualizadoEm         time.Time
		resolvidoEm                    sql.NullTime
	)

	err := row.Scan(
		&id, &idReferencia, &tipo, &payload, &status,
		&tentativas, &maxTentativas, &ultimoErro,
		&proximaTentativa, &criadoEm, &atualizadoEm, &resolvidoEm,
	)
	if err == sql.ErrNoRows {
		return nil, dlq.ErrMensagemNaoEncontrada
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao escanear mensagem da DLQ: %w", err)
	}

	return r.reconstruir(id, idReferencia, tipo, payload, status,
		tentativas, maxTentativas, ultimoErro,
		proximaTentativa, criadoEm, atualizadoEm, resolvidoEm), nil
}

// escanearMensagens lê múltiplas mensagens a partir de *sql.Rows.
func (r *RepositorioDLQ) escanearMensagens(rows *sql.Rows) ([]*dlq.Mensagem, error) {
	var mensagens []*dlq.Mensagem

	for rows.Next() {
		var (
			id, idReferencia, tipo, status string
			payload                        []byte
			tentativas, maxTentativas      int
			ultimoErro                     sql.NullString
			proximaTentativa               sql.NullTime
			criadoEm, atualizadoEm         time.Time
			resolvidoEm                    sql.NullTime
		)

		if err := rows.Scan(
			&id, &idReferencia, &tipo, &payload, &status,
			&tentativas, &maxTentativas, &ultimoErro,
			&proximaTentativa, &criadoEm, &atualizadoEm, &resolvidoEm,
		); err != nil {
			return nil, fmt.Errorf("erro ao escanear mensagem da DLQ: %w", err)
		}

		mensagens = append(mensagens, r.reconstruir(
			id, idReferencia, tipo, payload, status,
			tentativas, maxTentativas, ultimoErro,
			proximaTentativa, criadoEm, atualizadoEm, resolvidoEm,
		))
	}

	return mensagens, rows.Err()
}

// reconstruir monta o agregado Mensagem a partir dos dados lidos do banco.
func (r *RepositorioDLQ) reconstruir(
	id, idReferencia, tipo string,
	payload []byte,
	status string,
	tentativas, maxTentativas int,
	ultimoErro sql.NullString,
	proximaTentativa sql.NullTime,
	criadoEm, atualizadoEm time.Time,
	resolvidoEm sql.NullTime,
) *dlq.Mensagem {
	erro := ""
	if ultimoErro.Valid {
		erro = ultimoErro.String
	}

	var proximaTentativaPtr *time.Time
	if proximaTentativa.Valid {
		t := proximaTentativa.Time.UTC()
		proximaTentativaPtr = &t
	}

	var resolvidoEmPtr *time.Time
	if resolvidoEm.Valid {
		t := resolvidoEm.Time.UTC()
		resolvidoEmPtr = &t
	}

	return dlq.Reconstituir(
		id, idReferencia, tipo,
		payload,
		dlq.Status(status),
		tentativas, maxTentativas,
		erro,
		proximaTentativaPtr,
		criadoEm.UTC(),
		atualizadoEm.UTC(),
		resolvidoEmPtr,
	)
}
