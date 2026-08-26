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