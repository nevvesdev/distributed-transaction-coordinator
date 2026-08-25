package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/domain/transaction"
	domainshared "github.com/nevvesdev/distributed-transaction-coordinator/internal/shared/domain"
)

// RepositorioTransacao implementa transaction.Repository usando MySQL.
type RepositorioTransacao struct {
	db *sql.DB
}

// NovoRepositorioTransacao cria uma nova instância do repositório.
func NovoRepositorioTransacao(db *sql.DB) *RepositorioTransacao {
	return &RepositorioTransacao{db: db}
}

// Salvar persiste uma nova transação no banco de dados.
func (r *RepositorioTransacao) Salvar(ctx context.Context, t *transaction.Transaction) error {
	payloadJSON, err := json.Marshal(t.Payload())
	if err != nil {
		return fmt.Errorf("erro ao serializar payload: %w", err)
	}

	query := `
		INSERT INTO transacoes
			(id, status, payload, timeout_segundos, chave_idem, criado_em, atualizado_em, expirado_em)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?)
	`

	var expiradoEm interface{}
	if t.ExpiradoEm() != nil {
		expiradoEm = formatarTempo(*t.ExpiradoEm())
	}

	_, err = r.db.ExecContext(ctx, query,
		t.ID(),
		string(t.Status()),
		payloadJSON,
		int64(t.Timeout().Seconds()),
		t.ChaveIdem(),
		formatarTempo(t.CriadoEm()),
		formatarTempo(t.AtualizadoEm()),
		expiradoEm,
	)
	if err != nil {
		return fmt.Errorf("erro ao salvar transação: %w", err)
	}

	return nil
}

// Atualizar persiste as mudanças de estado de uma transação existente.
func (r *RepositorioTransacao) Atualizar(ctx context.Context, t *transaction.Transaction) error {
	query := `
		UPDATE transacoes
		SET status = ?, atualizado_em = ?, expirado_em = ?
		WHERE id = ?
	`

	var expiradoEm interface{}
	if t.ExpiradoEm() != nil {
		expiradoEm = formatarTempo(*t.ExpiradoEm())
	}

	resultado, err := r.db.ExecContext(ctx, query,
		string(t.Status()),
		formatarTempo(t.AtualizadoEm()),
		expiradoEm,
		t.ID(),
	)
	if err != nil {
		return fmt.Errorf("erro ao atualizar transação: %w", err)
	}

	linhas, err := resultado.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar linhas afetadas: %w", err)
	}
	if linhas == 0 {
		return transaction.ErrTransacaoNaoEncontrada
	}

	return nil
}

// BuscarPorID retorna uma transação pelo seu identificador único.
func (r *RepositorioTransacao) BuscarPorID(ctx context.Context, id string) (*transaction.Transaction, error) {
	query := `
		SELECT id, status, payload, timeout_segundos, chave_idem,
		       criado_em, atualizado_em, expirado_em
		FROM transacoes
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.escanearTransacao(row)
}

// BuscarPorChaveIdem retorna uma transação pela chave de idempotência.
func (r *RepositorioTransacao) BuscarPorChaveIdem(ctx context.Context, chave string) (*transaction.Transaction, error) {
	query := `
		SELECT id, status, payload, timeout_segundos, chave_idem,
		       criado_em, atualizado_em, expirado_em
		FROM transacoes
		WHERE chave_idem = ?
	`

	row := r.db.QueryRowContext(ctx, query, chave)
	return r.escanearTransacao(row)
}

// ListarExpiradas retorna transações que ultrapassaram o tempo limite e ainda não foram finalizadas.
func (r *RepositorioTransacao) ListarExpiradas(ctx context.Context) ([]*transaction.Transaction, error) {
	query := `
		SELECT id, status, payload, timeout_segundos, chave_idem,
		       criado_em, atualizado_em, expirado_em
		FROM transacoes
		WHERE expirado_em < ?
		  AND status NOT IN ('CONFIRMADA', 'ABORTADA', 'EXPIRADA')
	`

	rows, err := r.db.QueryContext(ctx, query, formatarTempo(time.Now().UTC()))
	if err != nil {
		return nil, fmt.Errorf("erro ao listar transações expiradas: %w", err)
	}
	defer rows.Close()

	var transacoes []*transaction.Transaction
	for rows.Next() {
		t, err := r.escanearTransacaoRows(rows)
		if err != nil {
			return nil, err
		}
		transacoes = append(transacoes, t)
	}

	return transacoes, rows.Err()
}

// escanearTransacao lê uma transação a partir de um *sql.Row.
func (r *RepositorioTransacao) escanearTransacao(row *sql.Row) (*transaction.Transaction, error) {
	var (
		id              string
		status          string
		payloadJSON     []byte
		timeoutSegundos int64
		chaveIdem       string
		criadoEm        time.Time
		atualizadoEm    time.Time
		expiradoEm      sql.NullTime
	)

	err := row.Scan(&id, &status, &payloadJSON, &timeoutSegundos, &chaveIdem,
		&criadoEm, &atualizadoEm, &expiradoEm)
	if err == sql.ErrNoRows {
		return nil, transaction.ErrTransacaoNaoEncontrada
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao escanear transação: %w", err)
	}

	return r.reconstruir(id, status, payloadJSON, timeoutSegundos, chaveIdem,
		criadoEm, atualizadoEm, expiradoEm)
}

// escanearTransacaoRows lê uma transação a partir de *sql.Rows.
func (r *RepositorioTransacao) escanearTransacaoRows(rows *sql.Rows) (*transaction.Transaction, error) {
	var (
		id              string
		status          string
		payloadJSON     []byte
		timeoutSegundos int64
		chaveIdem       string
		criadoEm        time.Time
		atualizadoEm    time.Time
		expiradoEm      sql.NullTime
	)

	err := rows.Scan(&id, &status, &payloadJSON, &timeoutSegundos, &chaveIdem,
		&criadoEm, &atualizadoEm, &expiradoEm)
	if err != nil {
		return nil, fmt.Errorf("erro ao escanear transação: %w", err)
	}

	return r.reconstruir(id, status, payloadJSON, timeoutSegundos, chaveIdem,
		criadoEm, atualizadoEm, expiradoEm)
}

// reconstruir monta o agregado Transaction a partir dos dados lidos do banco.
func (r *RepositorioTransacao) reconstruir(
	id, status string,
	payloadJSON []byte,
	timeoutSegundos int64,
	chaveIdem string,
	criadoEm, atualizadoEm time.Time,
	expiradoEm sql.NullTime,
) (*transaction.Transaction, error) {
	var payload map[string]string
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("erro ao desserializar payload: %w", err)
	}

	var expiradoEmPtr *time.Time
	if expiradoEm.Valid {
		t := expiradoEm.Time.UTC()
		expiradoEmPtr = &t
	}

	_ = domainshared.ErrIDInvalido // garante uso do pacote compartilhado

	return transaction.Reconstituir(
		id,
		transaction.Status(status),
		payload,
		nil,
		time.Duration(timeoutSegundos)*time.Second,
		chaveIdem,
		criadoEm.UTC(),
		atualizadoEm.UTC(),
		expiradoEmPtr,
	), nil
}
