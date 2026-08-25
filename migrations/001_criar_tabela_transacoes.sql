CREATE TABLE IF NOT EXISTS transacoes (
    id              VARCHAR(36)     NOT NULL,
    status          VARCHAR(20)     NOT NULL,
    payload         JSON            NOT NULL,
    timeout_segundos BIGINT         NOT NULL,
    chave_idem      VARCHAR(255)    NOT NULL,
    criado_em       DATETIME(6)     NOT NULL,
    atualizado_em   DATETIME(6)     NOT NULL,
    expirado_em     DATETIME(6)         NULL,

    CONSTRAINT pk_transacoes PRIMARY KEY (id),
    CONSTRAINT uq_transacoes_chave_idem UNIQUE (chave_idem),
    INDEX idx_transacoes_status (status),
    INDEX idx_transacoes_expirado_em (expirado_em)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;