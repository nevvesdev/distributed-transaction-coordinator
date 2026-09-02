package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// HealthHandler expõe o endpoint de health check detalhado da aplicação.
type HealthHandler struct {
	db          *sql.DB
	redisClient *goredis.Client
}

// NovoHealthHandler cria uma nova instância do handler de health check.
func NovoHealthHandler(db *sql.DB, redisClient *goredis.Client) *HealthHandler {
	return &HealthHandler{db: db, redisClient: redisClient}
}

// respostaHealth representa o corpo da resposta do health check.
type respostaHealth struct {
	Status    string                   `json:"status"`
	Timestamp string                   `json:"timestamp"`
	Servicos  map[string]statusServico `json:"servicos"`
}

// statusServico representa o status de um serviço dependente.
type statusServico struct {
	Status   string `json:"status"`
	Latencia string `json:"latencia,omitempty"`
	Erro     string `json:"erro,omitempty"`
}

// Verificar retorna o status de saúde da aplicação e suas dependências.
// GET /health
func (h *HealthHandler) Verificar(w http.ResponseWriter, r *http.Request) {
	ctx, cancelar := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancelar()

	servicos := make(map[string]statusServico)
	statusGeral := "saudavel"

	// verifica MySQL
	inicioMySQL := time.Now()
	if err := h.db.PingContext(ctx); err != nil {
		servicos["mysql"] = statusServico{
			Status: "degradado",
			Erro:   err.Error(),
		}
		statusGeral = "degradado"
	} else {
		servicos["mysql"] = statusServico{
			Status:   "saudavel",
			Latencia: time.Since(inicioMySQL).String(),
		}
	}

	// verifica Redis
	inicioRedis := time.Now()
	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		servicos["redis"] = statusServico{
			Status: "degradado",
			Erro:   err.Error(),
		}
		statusGeral = "degradado"
	} else {
		servicos["redis"] = statusServico{
			Status:   "saudavel",
			Latencia: time.Since(inicioRedis).String(),
		}
	}

	statusHTTP := http.StatusOK
	if statusGeral == "degradado" {
		statusHTTP = http.StatusServiceUnavailable
	}

	ResponderJSON(w, statusHTTP, respostaHealth{
		Status:    statusGeral,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Servicos:  servicos,
	})
}
