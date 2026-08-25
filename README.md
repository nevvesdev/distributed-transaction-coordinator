# Distributed Transaction Coordinator

Coordenador de transações distribuídas construído em Go com foco em confiabilidade, rastreabilidade e resiliência — aplicando padrões reais de sistemas financeiros distribuídos.

## Visão Geral

Este projeto implementa um orquestrador de transações distribuídas capaz de coordenar múltiplos participantes (microsserviços) garantindo consistência eventual ou forte, dependendo do protocolo utilizado.

## Funcionalidades

- **2-Phase Commit (2PC):** coordenação com timeout configurável e rollback automático
- **Saga Orquestrada:** fluxo de compensação gerenciado por um orquestrador central
- **Idempotência nativa:** deduplicação via chave de idempotência em todas as operações
- **Dead Letter Queue:** fila de reprocessamento com retry exponencial e limite de tentativas
- **Event Sourcing:** trilha de auditoria completa de todas as transições de estado
- **Distributed Locking:** locks atômicos via Redis para evitar condições de corrida

## Stack Técnica

| Camada | Tecnologia |
|--------|-----------|
| Linguagem | Go 1.26 |
| Roteador HTTP | chi v5 |
| Banco de dados | MySQL |
| Cache / Locks | Redis 7 |
| Métricas | Prometheus |
| Containerização | Docker Compose |