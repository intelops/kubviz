package application

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/client/pkg/clients"
	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/intelops/kubviz/client/pkg/storage"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/kelseyhightower/envconfig"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Application struct {
	Config   *config.Config
	conn     *clients.NATSContext
	dbClient clickhouse.DBInterface
}

func Start() *Application {
	log.Println("Client Application started...")

	ctx := context.Background()
	tracer := otel.Tracer("kubviz-client")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "Start")
	span.SetAttributes(attribute.String("start-app-client", "application"))
	defer span.End()

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
	c := cron.New()
	_, err = c.AddFunc("@daily", func() {
		if err := exportDataForTables(cfg); err != nil {
			log.Println("Error exporting data:", err)
		}
	})
	if err != nil {
		log.Fatal("Error adding cron job:", err)
	}

	// Listen for interrupt signals to stop the program
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Start the cron job scheduler
	c.Start()

	// Wait for an interrupt signal to stop the program
	<-interrupt

	// Stop the cron scheduler gracefully
	c.Stop()

	// Close the ClickHouse database connection
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

func exportDataForTables(cfg *config.Config) error {
	tables := []string{"events", "rakkess", "DeprecatedAPIs", "DeletedAPIs", "jfrogcontainerpush", "getall_resources", "outdated_images", "kubescore", "trivy_vul", "trivy_misconfig", "trivyimage", "dockerhubbuild", "azurecontainerpush", "quaycontainerpush", "trivysbom", "azure_devops", "github", "gitlab", "bitbucket", "gitea"}

	for _, tableName := range tables {
		err := storage.ExportExpiredData(tableName, cfg)
		if err != nil {
			log.Printf("Error exporting data for table %s: %v", tableName, err)
		} else {
			log.Printf("Export completed successfully for table %s.\n", tableName)
		}
	}

	return nil
}
