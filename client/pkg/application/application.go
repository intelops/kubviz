package application

import (
	"context"
	"log"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/client/pkg/clients"
	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/intelops/kubviz/client/pkg/storage"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/kelseyhightower/envconfig"
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

	tables := []string{"events", "rakkess", "DeprecatedAPIs", "DeletedAPIs", "jfrogcontainerpush", "getall_resources", "outdated_images", "kubescore", "trivy_vul", "trivy_misconfig", "trivyimage", "dockerhubbuild", "azurecontainerpush", "quaycontainerpush", "trivysbom", "azure_devops", "github", "gitlab", "bitbucket", "gitea"}

	for _, tableName := range tables {
		err := storage.ExportExpiredData(tableName, cfg)
		if err != nil {
			log.Printf("Error exporting data for table %s: %v", tableName, err)
		} else {
			log.Printf("Export completed successfully for table %s.\n", tableName)
		}
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
