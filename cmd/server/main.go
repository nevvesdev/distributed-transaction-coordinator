package main

import (
	"fmt"
	"log"

	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/config"
	"github.com/nevvesdev/distributed-transaction-coordinator/internal/infrastructure/persistence/mysql"
)

func main() {
	fmt.Println("Distributed Transaction Coordinator — inicializando...")

	cfg, err := config.Carregar()
	if err != nil {
		log.Fatalf("erro ao carregar configurações: %v", err)
	}

	db, err := mysql.NovaConexao(cfg.MySQL)
	if err != nil {
		log.Fatalf("erro ao conectar ao MySQL: %v", err)
	}
	defer db.Close()

	if err := mysql.ExecutarMigracoes(db); err != nil {
		log.Fatalf("erro ao executar migrações: %v", err)
	}

	fmt.Printf("servidor pronto na porta %s\n", cfg.Servidor.Porta)
}
