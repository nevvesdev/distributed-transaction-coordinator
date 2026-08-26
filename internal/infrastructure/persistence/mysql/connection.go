package mysql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/config"
)

// NovaConexao cria e configura um pool de conexões com o MySQL.
func NovaConexao(cfg config.ConfigMySQL) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC&multiStatements=true",
		cfg.Usuario,
		cfg.Senha,
		cfg.Host,
		cfg.Porta,
		cfg.Banco,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com MySQL: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxConexoesAbertas)
	db.SetMaxIdleConns(cfg.MaxConexoesInativas)
	db.SetConnMaxLifetime(cfg.TempoMaximoConexao)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao verificar conexão com MySQL: %w", err)
	}

	return db, nil
}

// ExecutarMigracoes aplica os scripts SQL de migração na ordem correta.
func ExecutarMigracoes(db *sql.DB) error {
	migracoes := []string{
		migracaoTransacoes,
		migracaoParticipantes,
		migracaoEventos,
		migracaoSagas,
		migracaoSagaSteps,
	}

	for _, migracao := range migracoes {
		if _, err := db.Exec(migracao); err != nil {
			return fmt.Errorf("erro ao executar migração: %w", err)
		}
	}

	return nil
}

// migracaoTransacoes contém o DDL da tabela de transações.
const migracaoTransacoes = `
CREATE TABLE IF NOT EXISTS transacoes (
    id               VARCHAR(36)  NOT NULL,
    status           VARCHAR(20)  NOT NULL,
    payload          JSON         NOT NULL,
    timeout_segundos BIGINT       NOT NULL,
    chave_idem       VARCHAR(255) NOT NULL,
    criado_em        DATETIME(6)  NOT NULL,
    atualizado_em    DATETIME(6)  NOT NULL,
    expirado_em      DATETIME(6)      NULL,

    CONSTRAINT pk_transacoes PRIMARY KEY (id),
    CONSTRAINT uq_transacoes_chave_idem UNIQUE (chave_idem),
    INDEX idx_transacoes_status (status),
    INDEX idx_transacoes_expirado_em (expirado_em)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// migracaoParticipantes contém o DDL da tabela de participantes.
const migracaoParticipantes = `
CREATE TABLE IF NOT EXISTS participantes (
    id           VARCHAR(36)  NOT NULL,
    id_transacao VARCHAR(36)  NOT NULL,
    endpoint     VARCHAR(500) NOT NULL,
    status       VARCHAR(20)  NOT NULL,
    tentativas   INT          NOT NULL DEFAULT 0,
    ultimo_erro  TEXT             NULL,
    criado_em    DATETIME(6)  NOT NULL,
    atualizado_em DATETIME(6) NOT NULL,

    CONSTRAINT pk_participantes PRIMARY KEY (id),
    CONSTRAINT fk_participantes_transacao
        FOREIGN KEY (id_transacao) REFERENCES transacoes(id)
        ON DELETE CASCADE,
    INDEX idx_participantes_id_transacao (id_transacao),
    INDEX idx_participantes_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// migracaoEventos contém o DDL da tabela de eventos de domínio.
const migracaoEventos = `
CREATE TABLE IF NOT EXISTS eventos_dominio (
    id          BIGINT       NOT NULL AUTO_INCREMENT,
    id_agregado VARCHAR(36)  NOT NULL,
    nome_evento VARCHAR(100) NOT NULL,
    payload     JSON         NOT NULL,
    ocorrido_em DATETIME(6)  NOT NULL,

    CONSTRAINT pk_eventos_dominio PRIMARY KEY (id),
    INDEX idx_eventos_id_agregado (id_agregado),
    INDEX idx_eventos_nome_evento (nome_evento),
    INDEX idx_eventos_ocorrido_em (ocorrido_em)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// formatarTempo formata um *time.Time para string compatível com MySQL DATETIME(6).
func formatarTempo(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05.000000")
}

// migracaoSagas contém o DDL da tabela de sagas.
const migracaoSagas = `
CREATE TABLE IF NOT EXISTS sagas (
    id            VARCHAR(36)  NOT NULL,
    id_transacao  VARCHAR(36)  NOT NULL,
    nome          VARCHAR(100) NOT NULL,
    status        VARCHAR(20)  NOT NULL,
    step_atual    INT          NOT NULL DEFAULT 0,
    criado_em     DATETIME(6)  NOT NULL,
    atualizado_em DATETIME(6)  NOT NULL,

    CONSTRAINT pk_sagas PRIMARY KEY (id),
    CONSTRAINT uq_sagas_id_transacao UNIQUE (id_transacao),
    INDEX idx_sagas_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// migracaoSagaSteps contém o DDL da tabela de steps da saga.
const migracaoSagaSteps = `
CREATE TABLE IF NOT EXISTS saga_steps (
    id               VARCHAR(36)  NOT NULL,
    id_saga          VARCHAR(36)  NOT NULL,
    nome             VARCHAR(100) NOT NULL,
    ordem            INT          NOT NULL,
    endpoint         VARCHAR(500) NOT NULL,
    endpoint_compen  VARCHAR(500) NOT NULL,
    status           VARCHAR(20)  NOT NULL,
    tentativas       INT          NOT NULL DEFAULT 0,
    ultimo_erro      TEXT             NULL,
    iniciado_em      DATETIME(6)      NULL,
    concluido_em     DATETIME(6)      NULL,

    CONSTRAINT pk_saga_steps PRIMARY KEY (id),
    CONSTRAINT fk_saga_steps_saga
        FOREIGN KEY (id_saga) REFERENCES sagas(id)
        ON DELETE CASCADE,
    INDEX idx_saga_steps_id_saga (id_saga),
    INDEX idx_saga_steps_ordem (ordem)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`
