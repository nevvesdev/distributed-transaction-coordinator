CREATE TABLE IF NOT EXISTS participantes (
    id              VARCHAR(36)     NOT NULL,
    id_transacao    VARCHAR(36)     NOT NULL,
    endpoint        VARCHAR(500)    NOT NULL,
    status          VARCHAR(20)     NOT NULL,
    tentativas      INT             NOT NULL DEFAULT 0,
    ultimo_erro     TEXT                NULL,
    criado_em       DATETIME(6)     NOT NULL,
    atualizado_em   DATETIME(6)     NOT NULL,

    CONSTRAINT pk_participantes PRIMARY KEY (id),
    CONSTRAINT fk_participantes_transacao
        FOREIGN KEY (id_transacao) REFERENCES transacoes(id)
        ON DELETE CASCADE,
    INDEX idx_participantes_id_transacao (id_transacao),
    INDEX idx_participantes_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;