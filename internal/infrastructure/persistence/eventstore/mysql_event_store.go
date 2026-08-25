package eventstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/shared/domain"
)

// MySQLEventStore implementa EventStore usando MySQL.
type MySQLEventStore struct {
	db *sql.DB
}

// NovoMySQLEventStore cria uma nova instância do event store.
func NovoMySQLEventStore(db *sql.DB) *MySQLEventStore {
	return &MySQLEventStore{db: db}
}

// Salvar persiste os eventos de domínio no banco de dados.
func (s *MySQLEventStore) Salvar(ctx context.Context, eventos []domain.DomainEvent) error {
	if len(eventos) == 0 {
		return nil
	}

	query := `
		INSERT INTO eventos_dominio (id_agregado, nome_evento, payload, ocorrido_em)
		VALUES (?, ?, ?, ?)
	`

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("erro ao iniciar transação para salvar eventos: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("erro ao preparar statement de eventos: %w", err)
	}
	defer stmt.Close()

	for _, evento := range eventos {
		payloadJSON, err := json.Marshal(evento)
		if err != nil {
			return fmt.Errorf("erro ao serializar evento '%s': %w", evento.NomeEvento(), err)
		}

		_, err = stmt.ExecContext(ctx,
			evento.IDAgregado(),
			evento.NomeEvento(),
			payloadJSON,
			evento.OcorridoEm().UTC().Format("2006-01-02 15:04:05.000000"),
		)
		if err != nil {
			return fmt.Errorf("erro ao persistir evento '%s': %w", evento.NomeEvento(), err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("erro ao confirmar transação de eventos: %w", err)
	}

	return nil
}

// ListarPorAgregado retorna todos os eventos de um agregado ordenados cronologicamente.
func (s *MySQLEventStore) ListarPorAgregado(ctx context.Context, idAgregado string) ([]RegistroEvento, error) {
	query := `
		SELECT id, id_agregado, nome_evento, payload, ocorrido_em
		FROM eventos_dominio
		WHERE id_agregado = ?
		ORDER BY ocorrido_em ASC, id ASC
	`

	rows, err := s.db.QueryContext(ctx, query, idAgregado)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar eventos do agregado: %w", err)
	}
	defer rows.Close()

	var registros []RegistroEvento
	for rows.Next() {
		var r RegistroEvento
		var ocorridoEm string

		if err := rows.Scan(&r.ID, &r.IDAgregado, &r.NomeEvento, &r.Payload, &ocorridoEm); err != nil {
			return nil, fmt.Errorf("erro ao escanear evento: %w", err)
		}

		r.OcorridoEm = ocorridoEm
		registros = append(registros, r)
	}

	return registros, rows.Err()
}
