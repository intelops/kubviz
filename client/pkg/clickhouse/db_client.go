package clickhouse

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/kuberhealthy/kuberhealthy/v2/pkg/health"
	"go.opentelemetry.io/otel"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/intelops/kubviz/gitmodels/dbstatement"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/opentelemetry"
)

type DBClient struct {
	splconn driver.Conn
	conn    *sql.DB
	conf    *config.Config
}
type DBInterface interface {
	InsertKuberhealthyMetrics(health.State)
	InsertRakeesMetrics(model.RakeesMetrics)
	InsertKetallEvent(model.Resource)
	InsertOutdatedEvent(model.CheckResultfinal)
	InsertDeprecatedAPI(model.DeprecatedAPI)
	InsertDeletedAPI(model.DeletedAPI)
	InsertKubvizEvent(model.Metrics)
	InsertGitEvent(string)
	InsertKubeScoreMetrics(model.KubeScoreRecommendations)
	InsertTrivyImageMetrics(metrics model.TrivyImage)
	InsertTrivySbomMetrics(metrics model.Sbom)
	InsertTrivyMetrics(metrics model.Trivy)
	RetriveKetallEvent() ([]model.Resource, error)
	RetriveOutdatedEvent() ([]model.CheckResultfinal, error)
	RetriveKubepugEvent() ([]model.Result, error)
	RetrieveKubvizEvent() ([]model.DbEvent, error)
	InsertContainerEventDockerHub(model.DockerHubBuild)
	InsertContainerEventAzure(model.AzureContainerPushEventPayload)
	InsertContainerEventQuay(model.QuayImagePushPayload)
	InsertContainerEventJfrog(model.JfrogContainerPushEventPayload)
	InsertContainerEventGithub(string)
	InsertGitCommon(metrics model.GitCommonAttribute, statement dbstatement.DBStatement) error
	Close()
}

func NewDBClient(conf *config.Config) (DBInterface, *sql.DB, error) {
	ctx := context.Background()
	var connOptions clickhouse.Options

	if conf.ClickHouseUsername != "" && conf.ClickHousePassword != "" {
		fmt.Println("Using provided username and password")
		connOptions = clickhouse.Options{
			Addr:  []string{fmt.Sprintf("%s:%d", conf.DBAddress, conf.DbPort)},
			Debug: true,
			Auth: clickhouse.Auth{
				Username: conf.ClickHouseUsername,
				Password: conf.ClickHousePassword,
			},
			Debugf: func(format string, v ...interface{}) {
				fmt.Printf(format, v...)
			},
			Settings: clickhouse.Settings{
				"allow_experimental_object_type": 1,
			},
		}
		fmt.Printf("Connecting to ClickHouse using username and password")
	} else {
		fmt.Println("Using connection without username and password")
		connOptions = clickhouse.Options{
			Addr:  []string{fmt.Sprintf("%s:%d", conf.DBAddress, conf.DbPort)},
			Debug: true,
			Debugf: func(format string, v ...interface{}) {
				fmt.Printf(format, v...)
			},
			Settings: clickhouse.Settings{
				"allow_experimental_object_type": 1,
			},
		}
		fmt.Printf("Connecting to ClickHouse  without  usename and password")

	}

	splconn, err := clickhouse.Open(&connOptions)
	if err != nil {
		return nil, nil, err
	}

	if err := splconn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println("Authentication error:", err) // Print the error message here
		}
		return nil, nil, err
	}

	var connOption clickhouse.Options

	if conf.ClickHouseUsername != "" && conf.ClickHousePassword != "" {
		fmt.Println("Using provided username and password")
		connOption = clickhouse.Options{
			Addr:  []string{fmt.Sprintf("%s:%d", conf.DBAddress, conf.DbPort)},
			Debug: true,
			Auth: clickhouse.Auth{
				Username: conf.ClickHouseUsername,
				Password: conf.ClickHousePassword,
			},
		}
	} else {
		fmt.Println("Using connection without username and password")
		connOption = clickhouse.Options{
			Addr: []string{fmt.Sprintf("%s:%d", conf.DBAddress, conf.DbPort)},
		}
	}

	stdconn := clickhouse.OpenDB(&connOption)

	if err := stdconn.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println("Authentication error:", err)
		}
		return nil, nil, err
	}

	return &DBClient{splconn: splconn, conn: stdconn, conf: conf}, stdconn, nil
}

func (c *DBClient) InsertContainerEventAzure(pushEvent model.AzureContainerPushEventPayload) {

	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-container-azure")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertContainerEventAzure")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}

	stmt, err := tx.Prepare(string(InsertAzureContainerPushEvent))
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	defer stmt.Close()

	currentTime := time.Now().UTC()

	registryURL := pushEvent.Request.Host
	repositoryName := pushEvent.Target.Repository
	tag := pushEvent.Target.Tag

	if tag == "" {
		tag = "latest"
	}
	imageName := registryURL + "/" + repositoryName + ":" + tag
	size := pushEvent.Target.Size
	shaID := pushEvent.Target.Digest

	pushEventJSON, err := json.Marshal(pushEvent)
	if err != nil {
		log.Printf("Error while marshaling Azure Container Registry payload: %v", err)
		return
	}

	if _, err := stmt.Exec(
		registryURL,
		repositoryName,
		tag,
		imageName,
		string(pushEventJSON),
		size,
		shaID,
		currentTime,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertContainerEventQuay(pushEvent model.QuayImagePushPayload) {

	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-container-quay")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertContainerEventQuay")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}

	stmt, err := tx.Prepare(string(InsertQuayContainerPushEvent))
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	defer stmt.Close()

	currentTime := time.Now().UTC()

	dockerURL := pushEvent.DockerURL
	repository := pushEvent.Repository
	name := pushEvent.Name
	nameSpace := pushEvent.Namespace
	homePage := pushEvent.Homepage

	var tag string
	if pushEvent.UpdatedTags != nil {
		tag = strings.Join(pushEvent.UpdatedTags, ",")
	} else {
		tag = ""
	}

	pushEventJSON, err := json.Marshal(pushEvent)
	if err != nil {
		log.Printf("Error while marshaling Quay Container Registry payload: %v", err)
		return
	}

	if _, err := stmt.Exec(
		name,
		repository,
		nameSpace,
		dockerURL,
		homePage,
		tag,
		string(pushEventJSON),
		currentTime,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertContainerEventJfrog(pushEvent model.JfrogContainerPushEventPayload) {

	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-container-jfrog")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertContainerEventJfrog")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}

	stmt, err := tx.Prepare(string(InsertJfrogContainerPushEvent))
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	defer stmt.Close()

	currentTime := time.Now().UTC()

	registryURL := pushEvent.Data.Path
	repositoryName := pushEvent.Data.Name
	tag := pushEvent.Data.Tag

	if tag == "" {
		tag = "latest"
	}
	imageName := pushEvent.Data.ImageName
	size := pushEvent.Data.Size
	shaID := pushEvent.Data.SHA256

	pushEventJSON, err := json.Marshal(pushEvent)
	if err != nil {
		log.Printf("Error while marshaling Jfrog Container Registry payload: %v", err)
		return
	}

	if _, err := stmt.Exec(
		pushEvent.Domain,
		pushEvent.EventType,
		registryURL,
		repositoryName,
		shaID,
		size,
		imageName,
		tag,
		string(pushEventJSON),
		currentTime,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertRakeesMetrics(metrics model.RakeesMetrics) {

	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-rakees-metrics")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertRakeesMetrics")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}
	stmt, err := tx.Prepare(string(InsertRakees))
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	defer stmt.Close()

	currentTime := time.Now().UTC()

	if _, err := stmt.Exec(
		metrics.ClusterName,
		metrics.Name,
		metrics.Create,
		metrics.Delete,
		metrics.List,
		metrics.Update,
		currentTime,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertKetallEvent(metrics model.Resource) {

	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-ketall-event")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertKetallEvent")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}
	stmt, err := tx.Prepare(string(InsertKetall))
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	defer stmt.Close()

	currentTime := time.Now().UTC()

	if _, err := stmt.Exec(
		metrics.ClusterName,
		metrics.Namespace,
		metrics.Kind,
		metrics.Resource,
		metrics.Age,
		currentTime,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertKuberhealthyMetrics(metrics health.State) {
	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-kuberhealthy")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertKuberhealthy")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}

	stmt, err := tx.Prepare(InsertKuberhealthy)
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	for checkName, checkdata := range metrics.CheckDetails {
		ok := uint8(0)
		if checkdata.OK {
			ok = uint8(1)
		}
		errors := strings.Join(checkdata.Errors, ", ")
		kcd := model.KuberhealthyCheckDetail{
			CurrentUUID:      checkdata.CurrentUUID,
			CheckName:        checkName,
			OK:               ok,
			Errors:           errors,
			RunDuration:      checkdata.RunDuration,
			Namespace:        checkdata.Namespace,
			Node:             checkdata.Node,
			LastRun:          checkdata.LastRun.Time.UTC(),
			AuthoritativePod: checkdata.AuthoritativePod,
		}
		if _, err := stmt.Exec(
			kcd.CurrentUUID,
			kcd.CheckName,
			kcd.OK,
			kcd.Errors,
			kcd.RunDuration,
			kcd.Namespace,
			kcd.Node,
			kcd.LastRun,
			kcd.AuthoritativePod,
		); err != nil {
			log.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	stmt.Close()
}

func (c *DBClient) InsertOutdatedEvent(metrics model.CheckResultfinal) {
	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-outdated-event")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertOutdatedEvent")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}
	stmt, err := tx.Prepare(string(InsertOutdated))
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	defer stmt.Close()

	currentTime := time.Now().UTC()

	if _, err := stmt.Exec(
		metrics.ClusterName,
		metrics.Namespace,
		metrics.Pod,
		metrics.Image,
		metrics.Current,
		metrics.LatestVersion,
		metrics.VersionsBehind,
		currentTime,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertDeprecatedAPI(deprecatedAPI model.DeprecatedAPI) {
	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-depricated-event")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertDeprecatedAPI")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}
	stmt, err := tx.Prepare(string(InsertDepricatedApi))
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	defer stmt.Close()

	deprecated := uint8(0)
	if deprecatedAPI.Deprecated {
		deprecated = 1
	}

	currentTime := time.Now().UTC()

	for _, item := range deprecatedAPI.Items {
		if _, err := stmt.Exec(
			deprecatedAPI.ClusterName,
			item.ObjectName,
			deprecatedAPI.Description,
			deprecatedAPI.Kind,
			deprecated,
			item.Scope,
			currentTime,
		); err != nil {
			log.Fatal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertDeletedAPI(deletedAPI model.DeletedAPI) {
	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-deletedapi")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertDeletedAPI")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}
	stmt, err := tx.Prepare(string(InsertDeletedApi))
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	defer stmt.Close()
	deleted := uint8(0)
	if deletedAPI.Deleted {
		deleted = 1
	}

	currentTime := time.Now().UTC()

	for _, item := range deletedAPI.Items {
		if _, err := stmt.Exec(
			deletedAPI.ClusterName,
			item.ObjectName,
			deletedAPI.Group,
			deletedAPI.Kind,
			deletedAPI.Version,
			deletedAPI.Name,
			deleted,
			item.Scope,
			currentTime,
		); err != nil {
			log.Fatal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertKubvizEvent(metrics model.Metrics) {
	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-kubviz-event")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertKubvizEvent")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}
	stmt, err := tx.Prepare(string(InsertKubvizEvent))
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	defer stmt.Close()
	eventJson, _ := json.Marshal(metrics.Event)
	formattedFirstTimestamp := metrics.Event.FirstTimestamp.Time.UTC().Format("2006-01-02 15:04:05")
	formattedLastTimestamp := metrics.Event.LastTimestamp.Time.UTC().Format("2006-01-02 15:04:05")

	if _, err := stmt.Exec(
		metrics.ClusterName,
		string(metrics.Event.UID),
		time.Now().UTC(),
		metrics.Type,
		metrics.Event.Name,
		metrics.Event.Namespace,
		metrics.Event.InvolvedObject.Kind,
		metrics.Event.Message,
		metrics.Event.Reason,
		metrics.Event.Source.Host,
		string(eventJson),
		metrics.ImageName,
		formattedFirstTimestamp,
		formattedLastTimestamp,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
func (c *DBClient) InsertGitEvent(event string) {
	//opentelemetry
	ctx := context.Background()
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		tracer := otel.Tracer("insert-git-event")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertGitEvent")
		defer span.End()
	}

	batch, err := c.splconn.PrepareBatch(ctx, "INSERT INTO git_json")
	if err != nil {
		log.Fatal(err)
	}

	if err = batch.Append(event); err != nil {
		log.Fatal(err)
	}

	if err = batch.Send(); err != nil {
		log.Fatal(err)
	}
}
func (c *DBClient) InsertContainerEvent(event string) {

	//opentelemetry
	ctx := context.Background()
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		tracer := otel.Tracer("insert-container-event")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertContainerEvent")
		defer span.End()
	}

	batch, err := c.splconn.PrepareBatch(ctx, "INSERT INTO container_bridge")
	if err != nil {
		log.Fatal(err)
	}

	if err = batch.Append(event); err != nil {
		log.Fatal(err)
	}

	if err = batch.Send(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertKubeScoreMetrics(metrics model.KubeScoreRecommendations) {

	//opentelemetry
	ctx := context.Background()
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		tracer := otel.Tracer("insert-kubescore-event")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertKubeScoreMetrics")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(InsertKubeScore)
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	currentTime := time.Now().UTC()

	for _, result := range metrics.Report {
		for _, check := range result.Checks {
			for _, comments := range check.Comments {

				if _, err := stmt.Exec(
					metrics.ID,
					metrics.ClusterName,
					result.ObjectName,
					result.TypeMeta.Kind,
					result.TypeMeta.APIVersion,
					result.ObjectMeta.Name,
					result.ObjectMeta.Namespace,
					check.Check.TargetType,
					comments.Description,
					comments.Path,
					comments.Summary,
					result.FileName,
					int64(result.FileRow),
					currentTime,
				); err != nil {
					log.Println("Error while inserting KubeScore metrics:", err)
				}
			}

		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertTrivyMetrics(metrics model.Trivy) {
	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-trivy-metrics")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertTrivyMetrics")
		defer span.End()
	}

	for _, finding := range metrics.Report.Findings {
		for _, result := range finding.Results {
			for _, vulnerability := range result.Vulnerabilities {
				tx, err := c.conn.Begin()
				if err != nil {
					log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
				}
				stmt, err := tx.Prepare(InsertTrivyVul)
				if err != nil {
					log.Fatalf("error preparing statement: %v", err)
				}
				if _, err := stmt.Exec(
					metrics.ID,
					metrics.ClusterName,
					finding.Namespace,
					finding.Kind,
					finding.Name,
					vulnerability.VulnerabilityID,
					strings.Join(vulnerability.VendorIDs, " "),
					vulnerability.PkgID,
					vulnerability.PkgName,
					vulnerability.PkgPath,
					vulnerability.InstalledVersion,
					vulnerability.FixedVersion,
					vulnerability.Title,
					vulnerability.Severity,
					vulnerability.PublishedDate,
					vulnerability.LastModifiedDate,
				); err != nil {
					log.Fatal(err)
				}
				if err := tx.Commit(); err != nil {
					log.Fatal(err)
				}
				stmt.Close()
			}

			for _, misconfiguration := range result.Misconfigurations {
				tx, err := c.conn.Begin()
				if err != nil {
					log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
				}
				stmt, err := tx.Prepare(InsertTrivyMisconfig)
				if err != nil {
					log.Fatalf("error preparing statement: %v", err)
				}

				defer stmt.Close()

				currentTime := time.Now().UTC()

				if _, err := stmt.Exec(
					metrics.ID,
					metrics.ClusterName,
					finding.Namespace,
					finding.Kind,
					finding.Name,
					misconfiguration.ID,
					misconfiguration.AVDID,
					misconfiguration.Type,
					misconfiguration.Title,
					misconfiguration.Description,
					misconfiguration.Message,
					misconfiguration.Query,
					misconfiguration.Resolution,
					misconfiguration.Severity,
					string(misconfiguration.Status),
					currentTime,
				); err != nil {
					log.Fatal(err)
				}
				if err := tx.Commit(); err != nil {
					log.Fatal(err)
				}
			}
		}
	}

}
func (c *DBClient) InsertTrivyImageMetrics(metrics model.TrivyImage) {
	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-trivy-image")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertTrivyImageMetrics")
		defer span.End()
	}

	for _, result := range metrics.Report.Results {
		for _, vulnerability := range result.Vulnerabilities {
			tx, err := c.conn.Begin()
			if err != nil {
				log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
			}
			stmt, err := tx.Prepare(InsertTrivyImage)
			if err != nil {
				log.Fatalf("error preparing statement: %v", err)
			}

			if _, err := stmt.Exec(
				metrics.ID,
				metrics.ClusterName,
				metrics.Report.ArtifactName,
				// metrics.Report.Metadata.Size,
				// metrics.Report.Metadata.OS.Name,
				// metrics.Report.Metadata.ImageID,
				// strings.Join(metrics.Report.Metadata.DiffIDs, ","),
				// strings.Join(metrics.Report.Metadata.RepoTags, ","),
				// strings.Join(metrics.Report.Metadata.RepoDigests, ","),
				vulnerability.VulnerabilityID,
				vulnerability.PkgID,
				vulnerability.PkgName,
				vulnerability.InstalledVersion,
				vulnerability.FixedVersion,
				vulnerability.Title,
				vulnerability.Severity,
				vulnerability.PublishedDate,
				vulnerability.LastModifiedDate,
			); err != nil {
				log.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				log.Fatal(err)
			}
			stmt.Close()
		}

	}
}
func (c *DBClient) InsertTrivySbomMetrics(metrics model.Sbom) {
	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-trivy-sbom")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertTrivySbomMetrics")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}
	stmt, err := tx.Prepare(InsertTrivySbom)
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	data := metrics.Report
	bomComponents, ok := data["components"].([]interface{})
	if !ok {
		log.Println("error: components not found or not in expected format")
	}
	var componentName, bomRef, serialNumber, bomFormat, purl, componentType string
	var version int32
	// Iterate over the components to find the desired name
	for _, component := range bomComponents {
		componentMap, ok := component.(map[string]interface{})
		if !ok {
			log.Println("error: component not in expected format")
			continue
		}
		if name, ok := componentMap["name"].(string); ok {
			componentName = name
			break
		}
	}
	metadata, ok := data["metadata"].(map[string]interface{})
	if !ok {
		log.Println("error: metadata not found or not in expected format")
		return
	}

	component, ok := metadata["component"].(map[string]interface{})
	if !ok {
		log.Println("error: component not found or not in expected format")
		return
	}

	bomRef, _ = component["bom-ref"].(string)
	purl, _ = component["purl"].(string)
	componentType, _ = component["type"].(string)

	serialNumber, _ = data["serialNumber"].(string)
	version, _ = data["version"].(int32)
	bomFormat, _ = data["bomFormat"].(string)

	if _, err := stmt.Exec(
		metrics.ID,
		metrics.ClusterName,
		componentName,
		componentType,
		purl,
		bomRef,
		serialNumber,
		int32(version),
		bomFormat,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	stmt.Close()
}
func (c *DBClient) Close() {
	_ = c.conn.Close()
}

func DbUrl(conf *config.Config) string {
	return fmt.Sprintf("tcp://%s:%d?debug=true", conf.DBAddress, conf.DbPort)
}
func (c *DBClient) RetriveKetallEvent() ([]model.Resource, error) {
	rows, err := c.conn.Query("SELECT ClusterName, Namespace, Kind, Resource, Age FROM getall_resources")
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	defer rows.Close()
	var events []model.Resource
	for rows.Next() {
		var result model.Resource
		if err := rows.Scan(&result.ClusterName, &result.Namespace, &result.Kind, &result.Resource, &result.Age); err != nil {
			log.Printf("Error: %s", err)
			return nil, err
		}
		events = append(events, result)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	return events, nil
}
func (c *DBClient) RetriveOutdatedEvent() ([]model.CheckResultfinal, error) {
	rows, err := c.conn.Query("SELECT ClusterName, Namespace, Pod, CurrentImage, CurrentTag, LatestVersion, VersionsBehind FROM outdated_images")
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	defer rows.Close()
	var events []model.CheckResultfinal
	for rows.Next() {
		var result model.CheckResultfinal
		if err := rows.Scan(&result.ClusterName, &result.Namespace, &result.Pod, &result.Image, &result.Current, &result.LatestVersion, &result.VersionsBehind); err != nil {
			log.Printf("Error: %s", err)
			return nil, err
		}
		events = append(events, result)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	return events, nil
}
func (c *DBClient) RetriveKubepugEvent() ([]model.Result, error) {
	rows, err := c.conn.Query("SELECT result, cluster_name FROM deprecatedAPIs_and_deletedAPIs")
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	defer rows.Close()
	var events []model.Result
	for rows.Next() {
		var result model.Result
		if err := rows.Scan(&result, &result.ClusterName); err != nil {
			log.Printf("Error: %s", err)
			return nil, err
		}
		events = append(events, result)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	return events, nil
}
func (c *DBClient) RetrieveKubvizEvent() ([]model.DbEvent, error) {
	rows, err := c.conn.Query("SELECT ClusterName, Id, EventTime, OpType, Name, Namespace, Kind, Message, Reason, Host, Event, FirstTime, LastTime FROM events")
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	defer rows.Close()
	var events []model.DbEvent
	for rows.Next() {
		var dbEvent model.DbEvent
		if err := rows.Scan(&dbEvent.Cluster_name, &dbEvent.Id, &dbEvent.Event_time, &dbEvent.Op_type, &dbEvent.Name, &dbEvent.Namespace, &dbEvent.Kind, &dbEvent.Message, &dbEvent.Reason, &dbEvent.Host, &dbEvent.Event, &dbEvent.First_time, &dbEvent.Last_time); err != nil {
			log.Printf("Error: %s", err)
			return nil, err
		}
		eventJson, _ := json.Marshal(dbEvent)
		log.Printf("DB Event: %s", string(eventJson))
		events = append(events, dbEvent)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	return events, nil
}

func (c *DBClient) InsertContainerEventDockerHub(build model.DockerHubBuild) {

	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-container-dockerhub")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertContainerEventDockerHub")
		defer span.End()
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}
	stmt, err := tx.Prepare(string(InsertDockerHubBuild))

	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}
	defer stmt.Close()

	currentTime := time.Now().UTC()

	if _, err := stmt.Exec(
		build.PushedBy,
		build.ImageTag,
		build.RepositoryName,
		build.DateCreated,
		build.Owner,
		build.Event,
		currentTime,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertContainerEventGithub(event string) {
	//opentelemetry
	opentelconfig, errs := opentelemetry.GetConfigurations()
	if errs != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("insert-container-github")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "InsertContainerEventGithub")
		defer span.End()
	}

	var image model.GithubImage
	err := json.Unmarshal([]byte(event), &image)
	if err != nil {
		log.Printf("Unable to unmarshal the Github image details: %v", err)
		return
	}

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatal(err)
	}

	query := "INSERT INTO container_github (package_id, created_at, image_name, organisation, updated_at, visibility, sha_id, image_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(
		image.PackageId,
		image.CreatedAt,
		image.ImageName,
		image.Organisation,
		image.UpdatedAt,
		image.Visibility,
		image.ShaID,
		image.ImageId,
	); err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertGitCommon(metrics model.GitCommonAttribute, statement dbstatement.DBStatement) error {
	tx, err := c.conn.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(string(statement))
	if err != nil {
		return err
	}
	defer stmt.Close()

	currentTime := time.Now().UTC()

	if _, err := stmt.Exec(
		metrics.Author,
		metrics.GitProvider,
		metrics.CommitID,
		metrics.CommitUrl,
		metrics.EventType,
		metrics.RepoName,
		currentTime,
		metrics.Event,
	); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
