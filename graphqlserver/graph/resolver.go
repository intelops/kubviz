package graph

import (
	"database/sql"
	"log"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/kelseyhightower/envconfig"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB *sql.DB
}

func NewResolver() *Resolver {
	log.Println("Client Application started...")
	cfg := &config.Config{}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatalf("Could not parse env Config: %v", err)
	}
	_, db, err := clickhouse.NewDBClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return &Resolver{DB: db}
}
