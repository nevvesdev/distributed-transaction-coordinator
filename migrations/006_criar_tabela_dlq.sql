CREATE TABLE IF NOT EXISTS dlq_mensagens (
    id                VARCHAR(36)  NOT NULL,
    id_referencia     VARCHAR(36)  NOT NULL,
    tipo              VARCHAR(100) NOT NULL,
    payload           JSON         NOT NULL,
    status            VARCHAR(20)  NOT NULL,
    tentativas        INT          NOT NULL DEFAULT 0,
    max_tentativas    INT          NOT NULL DEFAULT 5,
    ultimo_erro       TEXT             NULL,
    proxima_tentativa DATETIME(6)      NULL,
    criado_em         DATETIME(6)  NOT NULL,
    atualizado_em     DATETIME(6)  NOT NULL,
    resolvido_em      DATETIME(6)      NULL,

    CONSTRAINT pk_dlq_mensagens PRIMARY KEY (id),
    INDEX idx_dlq_status (status),
    INDEX idx_dlq_id_referencia (id_referencia),
    INDEX idx_dlq_proxima_tentativa (proxima_tentativa)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;