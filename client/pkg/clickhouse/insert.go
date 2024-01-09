package clickhouse

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/intelops/kubviz/gitmodels/dbstatement"
	"github.com/intelops/kubviz/model"
)

func (c *DBClient) InsertContainerEventDockerHub(build model.DockerHubBuild) {
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

func (c *DBClient) InsertContainerEventAzure(pushEvent model.AzureContainerPushEventPayload) {

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

func (c *DBClient) InsertOutdatedEvent(metrics model.CheckResultfinal) {
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
func (c *DBClient) InsertTrivySbomMetrics(metrics model.SbomData) {
	log.Println("####started inserting value")

	tx, err := c.conn.Begin()
	if err != nil {
		log.Fatalf("error beginning transaction, clickhouse connection not available: %v", err)
	}
	stmt, err := tx.Prepare(InsertTrivySbom)
	if err != nil {
		log.Fatalf("error preparing statement: %v", err)
	}

	if _, err := stmt.Exec(
		metrics.ID,
		metrics.ClusterName,
		metrics.ComponentName,
		metrics.PackageName,
		metrics.PackageUrl,
		metrics.BomRef,
		metrics.SerialNumber,
		int32(metrics.CycloneDxVersion),
		metrics.BomFormat,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	stmt.Close()
}
