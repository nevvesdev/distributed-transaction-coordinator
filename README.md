# Distributed Transaction Coordinator

Coordenador de transações distribuídas construído em Go com foco em confiabilidade, rastreabilidade e resiliência — aplicando padrões reais de sistemas financeiros distribuídos.

---

## Visão Geral

Este projeto implementa um orquestrador de transações distribuídas capaz de coordenar múltiplos participantes (microsserviços) garantindo consistência eventual ou forte, dependendo do protocolo utilizado.

---

## Funcionalidades

- **2-Phase Commit (2PC):** coordenação com timeout configurável e rollback automático
- **Saga Orquestrada:** fluxo de compensação gerenciado por um orquestrador central
- **Idempotência nativa:** deduplicação via chave de idempotência em todas as operações
- **Dead Letter Queue:** fila de reprocessamento com retry exponencial e jitter
- **Event Sourcing:** trilha de auditoria completa de todas as transições de estado
- **Distributed Locking:** locks atômicos via Redis para evitar condições de corrida
- **Observabilidade:** métricas Prometheus e health check detalhado

---

## Stack Técnica

| Camada | Tecnologia |
|--------|-----------|
| Linguagem | Go 1.26 |
| Roteador HTTP | chi v5 |
| Banco de dados | MySQL |
| Cache / Locks | Redis 7 |
| Métricas | Prometheus |
| Containerização | Docker Compose |

---

## Arquitetura

```
Clean Architecture + Domain-Driven Design + CQRS

---

## Endpoints

### 2-Phase Commit

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/transacoes` | Cria uma nova transação distribuída |
| `GET` | `/transacoes/{id}` | Busca uma transação pelo ID |
| `POST` | `/transacoes/{id}/participantes` | Registra um participante |
| `GET` | `/transacoes/{id}/participantes` | Lista os participantes |
| `POST` | `/transacoes/{id}/prepare` | Inicia a fase de Prepare |
| `POST` | `/transacoes/{id}/commit` | Inicia a fase de Commit |
| `POST` | `/transacoes/{id}/abort` | Aborta a transação |

### Saga Orquestrada

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/sagas` | Cria e executa uma Saga Orquestrada |
| `GET` | `/sagas/{id}` | Busca uma Saga pelo ID |

### Dead Letter Queue

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/dlq` | Enfileira uma mensagem manualmente |
| `GET` | `/dlq/{id_referencia}` | Lista mensagens por referência |

### Audit Trail

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET` | `/audit/{id_agregado}` | Histórico de eventos do agregado |
| `GET` | `/audit/transacoes/{id}` | Detalhe de uma transação |
| `GET` | `/audit/transacoes/{id}/projecao` | Projeção via Event Sourcing |
| `GET` | `/audit/transacoes/{id}/participantes` | Participantes formatados |

### Observabilidade

| Método | Rota | Descrição |
|--------|------|-----------|
| `GET` | `/health` | Health check com status MySQL e Redis |
| `GET` | `/metrics` | Métricas Prometheus |