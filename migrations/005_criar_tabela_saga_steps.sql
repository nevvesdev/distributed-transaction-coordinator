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