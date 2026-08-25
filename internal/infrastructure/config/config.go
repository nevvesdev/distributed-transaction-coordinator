package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// agrupa todas as configurações da aplicação lidas de variáveis de ambiente.
type Config struct {
	Servidor  ConfigServidor
	MySQL     ConfigMySQL
	Redis     ConfigRedis
	Transacao ConfigTransacao
}

type ConfigServidor struct {
	Porta          string
	TimeoutLeitura time.Duration
	TimeoutEscrita time.Duration
}

type ConfigMySQL struct {
	Host                string
	Porta               string
	Banco               string
	Usuario             string
	Senha               string
	MaxConexoesAbertas  int
	MaxConexoesInativas int
	TempoMaximoConexao  time.Duration
}

type ConfigRedis struct {
	Host    string
	Porta   string
	Senha   string
	DB      int
	Timeout time.Duration
}

type ConfigTransacao struct {
	TimeoutPadrao      time.Duration
	TimeoutPrepare     time.Duration
	MaxTentativasDLQ   int
	IntervaloBaseRetry time.Duration
}

// Carregar lê as variáveis de ambiente e retorna a configuração da aplicação.
func Carregar() (*Config, error) {
	maxConexoesAbertas, err := strconv.Atoi(obterEnv("MYSQL_MAX_CONEXOES_ABERTAS", "25"))
	if err != nil {
		return nil, fmt.Errorf("MYSQL_MAX_CONEXOES_ABERTAS inválido: %w", err)
	}

	maxConexoesInativas, err := strconv.Atoi(obterEnv("MYSQL_MAX_CONEXOES_INATIVAS", "25"))
	if err != nil {
		return nil, fmt.Errorf("MYSQL_MAX_CONEXOES_INATIVAS inválido: %w", err)
	}

	redisDB, err := strconv.Atoi(obterEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("REDIS_DB inválido: %w", err)
	}

	maxTentativasDLQ, err := strconv.Atoi(obterEnv("TRANSACAO_MAX_TENTATIVAS_DLQ", "5"))
	if err != nil {
		return nil, fmt.Errorf("TRANSACAO_MAX_TENTATIVAS_DLQ inválido: %w", err)
	}

	tempoMaximoConexao, err := time.ParseDuration(obterEnv("MYSQL_TEMPO_MAXIMO_CONEXAO", "5m"))
	if err != nil {
		return nil, fmt.Errorf("MYSQL_TEMPO_MAXIMO_CONEXAO inválido: %w", err)
	}

	redisTimeout, err := time.ParseDuration(obterEnv("REDIS_TIMEOUT", "5s"))
	if err != nil {
		return nil, fmt.Errorf("REDIS_TIMEOUT inválido: %w", err)
	}

	timeoutPadrao, err := time.ParseDuration(obterEnv("TRANSACAO_TIMEOUT_PADRAO", "30s"))
	if err != nil {
		return nil, fmt.Errorf("TRANSACAO_TIMEOUT_PADRAO inválido: %w", err)
	}

	timeoutPrepare, err := time.ParseDuration(obterEnv("TRANSACAO_TIMEOUT_PREPARE", "10s"))
	if err != nil {
		return nil, fmt.Errorf("TRANSACAO_TIMEOUT_PREPARE inválido: %w", err)
	}

	intervaloBaseRetry, err := time.ParseDuration(obterEnv("TRANSACAO_INTERVALO_BASE_RETRY", "1s"))
	if err != nil {
		return nil, fmt.Errorf("TRANSACAO_INTERVALO_BASE_RETRY inválido: %w", err)
	}

	timeoutLeitura, err := time.ParseDuration(obterEnv("SERVIDOR_TIMEOUT_LEITURA", "15s"))
	if err != nil {
		return nil, fmt.Errorf("SERVIDOR_TIMEOUT_LEITURA inválido: %w", err)
	}

	timeoutEscrita, err := time.ParseDuration(obterEnv("SERVIDOR_TIMEOUT_ESCRITA", "15s"))
	if err != nil {
		return nil, fmt.Errorf("SERVIDOR_TIMEOUT_ESCRITA inválido: %w", err)
	}

	return &Config{
		Servidor: ConfigServidor{
			Porta:          obterEnv("SERVIDOR_PORTA", "8080"),
			TimeoutLeitura: timeoutLeitura,
			TimeoutEscrita: timeoutEscrita,
		},
		MySQL: ConfigMySQL{
			Host:                obterEnv("MYSQL_HOST", "localhost"),
			Porta:               obterEnv("MYSQL_PORT", "3306"),
			Banco:               obterEnv("MYSQL_DATABASE", "dtc"),
			Usuario:             obterEnv("MYSQL_USER", "dtc_usuario"),
			Senha:               obterEnv("MYSQL_PASSWORD", "dtc_senha"),
			MaxConexoesAbertas:  maxConexoesAbertas,
			MaxConexoesInativas: maxConexoesInativas,
			TempoMaximoConexao:  tempoMaximoConexao,
		},
		Redis: ConfigRedis{
			Host:    obterEnv("REDIS_HOST", "localhost"),
			Porta:   obterEnv("REDIS_PORT", "6379"),
			Senha:   obterEnv("REDIS_PASSWORD", "redis_senha"),
			DB:      redisDB,
			Timeout: redisTimeout,
		},
		Transacao: ConfigTransacao{
			TimeoutPadrao:      timeoutPadrao,
			TimeoutPrepare:     timeoutPrepare,
			MaxTentativasDLQ:   maxTentativasDLQ,
			IntervaloBaseRetry: intervaloBaseRetry,
		},
	}, nil
}

func obterEnv(chave, padrao string) string {
	if valor := os.Getenv(chave); valor != "" {
		return valor
	}
	return padrao
}
