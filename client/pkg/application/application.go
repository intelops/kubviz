package application

import (
	"context"
	"database/sql"
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

const (
	EventsTable             = "events"
	RakkessTable            = "rakkess"
	DeprecatedAPIsTable     = "DeprecatedAPIs"
	DeletedAPIsTable        = "DeletedAPIs"
	JfrogContainerPushTable = "jfrogcontainerpush"
	GetAllResourcesTable    = "getall_resources"
	OutdatedImagesTable     = "outdated_images"
	KubeScoreTable          = "kubescore"
	TrivyVulTable           = "trivy_vul"
	TrivyMisconfigTable     = "trivy_misconfig"
	TrivyImageTable         = "trivyimage"
	DockerHubBuildTable     = "dockerhubbuild"
	AzureContainerPushTable = "azurecontainerpush"
	QuayContainerPushTable  = "quaycontainerpush"
	TrivySBOMTable          = "trivysbom"
	AzureDevOpsTable        = "azure_devops"
	GitHubTable             = "github"
	GitLabTable             = "gitlab"
	BitbucketTable          = "bitbucket"
	GiteaTable              = "gitea"
)

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
	dbClient, conn, err := clickhouse.NewDBClient(cfg)
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
		if err := exportDataForTables(conn); err != nil {
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
func exportDataForTables(db *sql.DB) error {
	//pvcMountPath := "/mnt/client/kbz"
	tables := []string{
		EventsTable, RakkessTable, DeprecatedAPIsTable, DeletedAPIsTable, JfrogContainerPushTable, GetAllResourcesTable, OutdatedImagesTable, KubeScoreTable, TrivyVulTable, TrivyMisconfigTable, TrivyImageTable, DockerHubBuildTable, AzureContainerPushTable, QuayContainerPushTable, TrivySBOMTable, AzureDevOpsTable, GitHubTable, GitLabTable, BitbucketTable, GiteaTable,
	}
	for _, tableName := range tables {
		err := storage.ExportExpiredData(tableName, db)
		if err != nil {
			log.Printf("Error exporting data for table %s: %v", tableName, err)
		} else {
			log.Printf("Export completed successfully for table %s.\n", tableName)
		}
	}

	return nil
}
