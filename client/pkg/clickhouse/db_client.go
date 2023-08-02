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

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/intelops/kubviz/gitmodels/dbstatement"
	"github.com/intelops/kubviz/model"
)

type DBClient struct {
	splconn driver.Conn
	conn    *sql.DB
	conf    *config.Config
}
type DBInterface interface {
	InsertRakeesMetrics(model.RakeesMetrics)
	InsertKetallEvent(model.Resource)
	InsertOutdatedEvent(model.CheckResultfinal)
	InsertDeprecatedAPI(model.DeprecatedAPI)
	InsertDeletedAPI(model.DeletedAPI)
	InsertKubvizEvent(model.Metrics)
	InsertGitEvent(string)
	InsertKubeScoreMetrics(model.KubeScoreRecommendations)
	InsertTrivyImageMetrics(metrics model.TrivyImage)
	InsertTrivyMetrics(metrics model.Trivy)
	RetriveKetallEvent() ([]model.Resource, error)
	RetriveOutdatedEvent() ([]model.CheckResultfinal, error)
	RetriveKubepugEvent() ([]model.Result, error)
	RetrieveKubvizEvent() ([]model.DbEvent, error)
	InsertContainerEventDockerHub(model.DockerHubBuild)
	InsertContainerEventGithub(string)
	InsertGitCommon(metrics model.GitCommonAttribute, statement dbstatement.DBStatement) error
	Close()
}

func NewDBClient(conf *config.Config) (DBInterface, error) {
	ctx := context.Background()
	log.Println("Connecting to Clickhouse DB and creating schemas...")
	splconn, err := clickhouse.Open(&clickhouse.Options{
		Addr:  []string{fmt.Sprintf("%s:%d", conf.DBAddress, conf.DbPort)},
		Debug: true,
		Debugf: func(format string, v ...any) {
			fmt.Printf(format, v)
		},
		Settings: clickhouse.Settings{
			"allow_experimental_object_type": 1,
		},
	})
	if err != nil {
		return nil, err
	}

	if err := splconn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return nil, err
	}

	tables := []DBStatement{kubvizTable, rakeesTable, kubePugDepricatedTable, kubepugDeletedTable, ketallTable, trivyTableImage, outdateTable, clickhouseExperimental, containerDockerhubTable, containerGithubTable, kubescoreTable, trivyTableVul, trivyTableMisconfig, dockerHubBuildTable, DBStatement(dbstatement.AzureDevopsTable), DBStatement(dbstatement.GithubTable), DBStatement(dbstatement.GitlabTable), DBStatement(dbstatement.BitbucketTable), DBStatement(dbstatement.GiteaTable)}
	for _, table := range tables {
		if err = splconn.Exec(context.Background(), string(table)); err != nil {
			return nil, err
		}
	}
	stdconn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", conf.DBAddress, conf.DbPort)},
	})
	if err := stdconn.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return nil, err
	}
	return &DBClient{splconn: splconn, conn: stdconn, conf: conf}, nil
}

func (c *DBClient) InsertRakeesMetrics(metrics model.RakeesMetrics) {
	var (
		tx, _   = c.conn.Begin()
		stmt, _ = tx.Prepare(string(InsertRakees))
	)
	defer stmt.Close()
	if _, err := stmt.Exec(
		metrics.ClusterName,
		metrics.Name,
		metrics.Create,
		metrics.Delete,
		metrics.List,
		metrics.Update,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
func (c *DBClient) InsertKetallEvent(metrics model.Resource) {
	var (
		tx, _   = c.conn.Begin()
		stmt, _ = tx.Prepare(string(InsertKetall))
	)
	defer stmt.Close()
	if _, err := stmt.Exec(
		metrics.ClusterName,
		metrics.Namespace,
		metrics.Kind,
		metrics.Resource,
		metrics.Age,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
func (c *DBClient) InsertOutdatedEvent(metrics model.CheckResultfinal) {
	var (
		tx, _   = c.conn.Begin()
		stmt, _ = tx.Prepare(string(InsertOutdated))
	)
	defer stmt.Close()
	if _, err := stmt.Exec(
		metrics.ClusterName,
		metrics.Namespace,
		metrics.Pod,
		metrics.Image,
		metrics.Current,
		metrics.LatestVersion,
		metrics.VersionsBehind,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
func (c *DBClient) InsertDeprecatedAPI(deprecatedAPI model.DeprecatedAPI) {
	var (
		tx, _   = c.conn.Begin()
		stmt, _ = tx.Prepare(string(InsertDepricatedApi))
	)
	defer stmt.Close()

	deprecated := uint8(0)
	if deprecatedAPI.Deprecated {
		deprecated = 1
	}

	for _, item := range deprecatedAPI.Items {
		if _, err := stmt.Exec(
			deprecatedAPI.ClusterName,
			item.ObjectName,
			deprecatedAPI.Description,
			deprecatedAPI.Kind,
			deprecated,
			item.Scope,
		); err != nil {
			log.Fatal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
func (c *DBClient) InsertDeletedAPI(deletedAPI model.DeletedAPI) {
	var (
		tx, _   = c.conn.Begin()
		stmt, _ = tx.Prepare(string(InsertDeletedApi))
	)
	defer stmt.Close()
	deleted := uint8(0)
	if deletedAPI.Deleted {
		deleted = 1
	}

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
		); err != nil {
			log.Fatal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
func (c *DBClient) InsertKubvizEvent(metrics model.Metrics) {
	var (
		tx, _   = c.conn.Begin()
		stmt, _ = tx.Prepare(string(InsertKubvizEvent))
	)
	defer stmt.Close()
	eventJson, _ := json.Marshal(metrics.Event)
	if _, err := stmt.Exec(
		metrics.ClusterName,
		string(metrics.Event.UID),
		time.Now().Format("2006-01-02 15:04:05"),
		metrics.Type,
		metrics.Event.Name,
		metrics.Event.Namespace,
		metrics.Event.InvolvedObject.Kind,
		metrics.Event.Message,
		metrics.Event.Reason,
		metrics.Event.Source.Host,
		string(eventJson),
		metrics.Event.FirstTimestamp.Time.Format("2006-01-02 15:04:05"),
		metrics.Event.LastTimestamp.Time.Format("2006-01-02 15:04:05"),
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
func (c *DBClient) InsertGitEvent(event string) {
	ctx := context.Background()
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
	ctx := context.Background()
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
	var (
		tx, _   = c.conn.Begin()
		stmt, _ = tx.Prepare(InsertKubeScore)
	)
	defer stmt.Close()
	if _, err := stmt.Exec(
		metrics.ID,
		metrics.Namespace,
		metrics.ClusterName,
		metrics.Recommendations,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertTrivyMetrics(metrics model.Trivy) {
	for _, finding := range metrics.Report.Findings {
		for _, result := range finding.Results {
			for _, vulnerability := range result.Vulnerabilities {
				var (
					tx, _   = c.conn.Begin()
					stmt, _ = tx.Prepare(InsertTrivyVul)
				)
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
				var (
					tx, _   = c.conn.Begin()
					stmt, _ = tx.Prepare(InsertTrivyMisconfig)
				)
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

}
func (c *DBClient) InsertTrivyImageMetrics(metrics model.TrivyImage) {
	for _, result := range metrics.Report.Results {
		for _, vulnerability := range result.Vulnerabilities {
			var (
				tx, _   = c.conn.Begin()
				stmt, _ = tx.Prepare(InsertTrivyImage)
			)
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
	var (
		tx, _   = c.conn.Begin()
		stmt, _ = tx.Prepare(string(InsertDockerHubBuild))
	)
	defer stmt.Close()
	if _, err := stmt.Exec(
		build.PushedBy,
		build.ImageTag,
		build.RepositoryName,
		build.DateCreated,
		build.Owner,
		build.Event,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (c *DBClient) InsertContainerEventGithub(event string) {
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
	if _, err := stmt.Exec(
		metrics.Author,
		metrics.GitProvider,
		metrics.CommitID,
		metrics.CommitUrl,
		metrics.EventType,
		metrics.RepoName,
		metrics.TimeStamp,
		metrics.Event,
	); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
