package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/config"
)

// Servidor encapsula o servidor HTTP da aplicação.
type Servidor struct {
	httpServer *http.Server
}

// NovoServidor cria e configura o servidor HTTP com os timeouts definidos em config.
func NovoServidor(cfg config.ConfigServidor, router http.Handler) *Servidor {
	return &Servidor{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", cfg.Porta),
			Handler:      router,
			ReadTimeout:  cfg.TimeoutLeitura,
			WriteTimeout: cfg.TimeoutEscrita,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Iniciar coloca o servidor em modo de escuta em uma goroutine dedicada.
func (s *Servidor) Iniciar() {
	go func() {
		log.Printf("servidor HTTP ouvindo em %s", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("erro ao iniciar servidor HTTP: %v", err)
		}
	}()
}

// Desligar encerra o servidor HTTP de forma graciosa respeitando o contexto.
func (s *Servidor) Desligar(ctx context.Context) error {
	log.Println("encerrando servidor HTTP...")
	return s.httpServer.Shutdown(ctx)
}
