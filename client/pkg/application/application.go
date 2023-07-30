package application

import (
	"log"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/client/pkg/clients"
	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/kelseyhightower/envconfig"
)

type Application struct {
	Config   *config.Config
	conn     *clients.NATSContext
	dbClient clickhouse.DBInterface
}

func Start() *Application {
	log.Println("Client Application started...")
	cfg := &config.Config{}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatalf("Could not parse env Config: %v", err)
	}
	dbClient, err := clickhouse.NewDBClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	// Connect to NATS
	natsContext, err := clients.NewNATSContext(cfg, dbClient)
	if err != nil {
		log.Fatal("Error establishing connection to NATS:", err)
	}
	return &Application{
		Config:   cfg,
		conn:     natsContext,
		dbClient: dbClient,
	}
}

func (app *Application) Close() {
	log.Printf("Closing the service gracefully")
	app.conn.Close()
	app.dbClient.Close()
}
