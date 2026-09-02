package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metricas centraliza todos os contadores e histogramas expostos ao Prometheus.
type Metricas struct {
	TransacoesIniciadas   prometheus.Counter
	TransacoesConfirmadas prometheus.Counter
	TransacoesAbortadas   prometheus.Counter
	TransacoesExpiradas   prometheus.Counter
	SagasExecutadas       prometheus.Counter
	SagasConcluidas       prometheus.Counter
	SagasCompensadas      prometheus.Counter
	MensagensDLQ          prometheus.Counter
	MensagensResolvidas   prometheus.Counter
	MensagensDescartadas  prometheus.Counter
	DuracaoRequisicao     *prometheus.HistogramVec
	OperacoesAtivas       prometheus.Gauge
}

// NovasMetricas registra e retorna todas as métricas da aplicação.
func NovasMetricas() *Metricas {
	return &Metricas{
		TransacoesIniciadas: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dtc_transacoes_iniciadas_total",
			Help: "Total de transações distribuídas iniciadas.",
		}),
		TransacoesConfirmadas: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dtc_transacoes_confirmadas_total",
			Help: "Total de transações confirmadas com sucesso.",
		}),
		TransacoesAbortadas: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dtc_transacoes_abortadas_total",
			Help: "Total de transações abortadas.",
		}),
		TransacoesExpiradas: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dtc_transacoes_expiradas_total",
			Help: "Total de transações expiradas por timeout.",
		}),
		SagasExecutadas: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dtc_sagas_executadas_total",
			Help: "Total de sagas iniciadas.",
		}),
		SagasConcluidas: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dtc_sagas_concluidas_total",
			Help: "Total de sagas concluídas com sucesso.",
		}),
		SagasCompensadas: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dtc_sagas_compensadas_total",
			Help: "Total de sagas compensadas após falha.",
		}),
		MensagensDLQ: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dtc_dlq_mensagens_total",
			Help: "Total de mensagens enfileiradas na DLQ.",
		}),
		MensagensResolvidas: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dtc_dlq_mensagens_resolvidas_total",
			Help: "Total de mensagens resolvidas pelo worker DLQ.",
		}),
		MensagensDescartadas: promauto.NewCounter(prometheus.CounterOpts{
			Name: "dtc_dlq_mensagens_descartadas_total",
			Help: "Total de mensagens descartadas após atingir limite de tentativas.",
		}),
		DuracaoRequisicao: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "dtc_duracao_requisicao_segundos",
			Help:    "Duração das requisições HTTP em segundos.",
			Buckets: prometheus.DefBuckets,
		}, []string{"metodo", "rota", "status"}),
		OperacoesAtivas: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "dtc_operacoes_ativas",
			Help: "Número de operações em andamento no momento.",
		}),
	}
}

// Handler retorna o handler HTTP para exposição das métricas ao Prometheus.
func Handler() http.Handler {
	return promhttp.Handler()
}
