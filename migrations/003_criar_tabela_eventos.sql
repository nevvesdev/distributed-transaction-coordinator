CREATE TABLE IF NOT EXISTS eventos_dominio (
    id              BIGINT          NOT NULL AUTO_INCREMENT,
    id_agregado     VARCHAR(36)     NOT NULL,
    nome_evento     VARCHAR(100)    NOT NULL,
    payload         JSON            NOT NULL,
    ocorrido_em     DATETIME(6)     NOT NULL,

    CONSTRAINT pk_eventos_dominio PRIMARY KEY (id),
    INDEX idx_eventos_id_agregado (id_agregado),
    INDEX idx_eventos_nome_evento (nome_evento),
    INDEX idx_eventos_ocorrido_em (ocorrido_em)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;