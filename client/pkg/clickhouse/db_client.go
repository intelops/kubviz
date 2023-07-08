package clickhouse

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go"
	"github.com/kube-tarian/kubviz/client/pkg/config"
	"github.com/kube-tarian/kubviz/model"
)

type DBClient struct {
	conn *sql.DB
	conf *config.Config
}
type DBInterface interface {
	InsertRakeesMetrics(model.RakeesMetrics)
	InsertKetallEvent(model.Resource)
	InsertOutdatedEvent(model.CheckResultfinal)
	InsertDeprecatedAPI(model.DeprecatedAPI)
	InsertDeletedAPI(model.DeletedAPI)
	InsertKubvizEvent(model.Metrics)
	InsertGitEvent(string)
	InsertContainerEvent(string)
	RetriveKetallEvent() ([]model.Resource, error)
	RetriveOutdatedEvent() ([]model.CheckResultfinal, error)
	RetriveKubepugEvent() ([]model.Result, error)
	RetrieveKubvizEvent() ([]model.DbEvent, error)
	Close()
}

func NewDBClient(conf *config.Config) (DBInterface, error) {
	log.Println("Connecting to Clickhouse DB and creating schemas...")
	conn, err := sql.Open("clickhouse", DbUrl(conf))
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return nil, err
	}
	tables := []DBStatement{kubvizTable, rakeesTable, kubePugDepricatedTable, kubepugDeletedTable, ketallTable, outdateTable, clickhouseExperimental, containerTable, gitTable}
	for _, table := range tables {
		if _, err = conn.Exec(string(table)); err != nil {
			return nil, err
		}
	}
	return &DBClient{conn: conn, conf: conf}, nil
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
			deprecatedAPI.Description,
			deprecatedAPI.Kind,
			deprecated,
			item.Scope,
			item.ObjectName,
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
			deletedAPI.Group,
			deletedAPI.Kind,
			deletedAPI.Version,
			deletedAPI.Name,
			deleted,
			item.Scope,
			item.ObjectName,
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
		metrics.Event.UID,
		metrics.Type,
		metrics.Event.Name,
		metrics.Event.Namespace,
		metrics.Event.InvolvedObject.Kind,
		metrics.Event.Message,
		metrics.Event.Reason,
		metrics.Event.Source.Host,
		string(eventJson),
		metrics.Event.FirstTimestamp.Time,
		metrics.Event.LastTimestamp.Time,
		time.Now(),
		metrics.ClusterName,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
func (c *DBClient) InsertGitEvent(event string) {
	var (
		tx, _   = c.conn.Begin()
		stmt, _ = tx.Prepare(fmt.Sprintf("INSERT INTO git_json FORMAT JSONAsObject %v", event))
	)
	defer stmt.Close()

	if _, err := stmt.Exec(); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
func (c *DBClient) InsertContainerEvent(event string) {
	var (
		tx, _   = c.conn.Begin()
		stmt, _ = tx.Prepare(fmt.Sprintf("INSERT INTO container_bridge FORMAT JSONAsObject %v", event))
	)
	defer stmt.Close()
	if _, err := stmt.Exec(); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
func (c *DBClient) Close() {
	_ = c.conn.Close()
}

func DbUrl(conf *config.Config) string {
	return fmt.Sprintf("tcp://%s:%d?debug=true", conf.DBAddress, conf.DbPort)
}
func (c *DBClient) RetriveKetallEvent() ([]model.Resource, error) {
	rows, err := c.conn.Query("SELECT Cluster_Name, Namespace, Kind, Resource, Age FROM getall_resources")
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
	rows, err := c.conn.Query("SELECT Cluster_Name, Namespace, Pod, Current_Image, Current_Tag, Latest_Version, Versions_Behind FROM outdated_images")
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
	rows, err := c.conn.Query("SELECT id, op_type, name, namespace, kind, message, reason, host, event, first_time, last_time, event_time, cluster_name FROM events")
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	defer rows.Close()
	var events []model.DbEvent
	for rows.Next() {
		var dbEvent model.DbEvent
		if err := rows.Scan(&dbEvent.Id, &dbEvent.Op_type, &dbEvent.Name, &dbEvent.Namespace, &dbEvent.Kind, &dbEvent.Message, &dbEvent.Host, &dbEvent.Event, &dbEvent.First_time, &dbEvent.Last_time, &dbEvent.Event_time, &dbEvent.Cluster_name); err != nil {
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
