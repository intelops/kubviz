package clickhouse

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ClickHouse/clickhouse-go"
	"github.com/kube-tarian/kubviz/model"
)

func GetClickHouseConnection(url string) (*sql.DB, error) {
	connect, err := sql.Open("clickhouse", url)
	//connect, err := sql.Open("clickhouse", "tcp://kubviz-client-clickhouse:9000?debug=true")
	if err != nil {
		log.Fatal(err)
	}
	if err := connect.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		} else {
			fmt.Println(err)
		}
		return nil, err
	}

	return connect, nil
}

func CreateSchema(connect *sql.DB) {
	_, err := connect.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id           UUID,
			op_type      String,
			name         String,
			namespace    String,
			kind         String,
			message      String,
			reason       String,
			host         String,
			event        String,
			first_time   DateTime,
			last_time    DateTime,
			event_time   DateTime,
			cluster_name String
		) engine=File(TabSeparated)
	`)

	if err != nil {
		log.Fatal(err)
	}
}

func CreateKubePugSchema(connect *sql.DB) {
	_, err := connect.Exec(`
		CREATE TABLE IF NOT EXISTS deprecatedAPIs_and_deletedAPIs (
			result String,
			cluster_name String
        ) engine=File(TabSeparated)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateKetallSchema(connect *sql.DB) {
	_, err := connect.Exec(`
		CREATE TABLE IF NOT EXISTS getall_resources (
			resource String,
			namespace String,
			age String,
			cluster_name String
        ) engine=File(TabSeparated)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateOutdatedSchema(connect *sql.DB) {
	_, err := connect.Exec(`
	    CREATE TABLE IF NOT EXISTS outdated_images (
		    current_image String,
			current_tag String,
			latest_version String,
			versions_behind Int64,
			cluster_name String
	    ) engine=File(TabSeparated)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateKubeScoreSchema(connect *sql.DB) {
	_, err := connect.Exec(`
	    CREATE TABLE IF NOT EXISTS kubescore (
		    id UUID,
			namespace String,
			cluster_name String,
			recommendations String
	    ) engine=File(TabSeparated)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func InsertKetallEvent(connect *sql.DB, metrics model.Resource) {
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO getall_resources (resource, namespace, age, cluster_name) VALUES (?, ?, ?, ?)")
	)
	defer stmt.Close()
	if _, err := stmt.Exec(
		metrics.Resource,
		metrics.Namespace,
		metrics.Age,
		metrics.ClusterName,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func InsertOutdatedEvent(connect *sql.DB, metrics model.CheckResultfinal) {
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO outdated_images (current_image, current_tag, latest_version, versions_behind, cluster_name) VALUES (?, ?, ?, ?, ?)")
	)
	defer stmt.Close()
	if _, err := stmt.Exec(
		metrics.Image,
		metrics.Current,
		metrics.LatestVersion,
		metrics.VersionsBehind,
		metrics.ClusterName,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func InsertKubepugEvent(connect *sql.DB, metrics model.Result) {
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO deprecatedAPIs_and_deletedAPIs (result, cluster_name) VALUES (?, ?)")
	)
	defer stmt.Close()
	eventJson, _ := json.Marshal(metrics)
	if _, err := stmt.Exec(
		string(eventJson),
		metrics.ClusterName,
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func InsertEvent(connect *sql.DB, metrics model.Metrics) {
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO events (id, op_type, name, namespace, kind, message, reason, host, event, first_time, last_time, event_time, cluster_name) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
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

func InsertKubeScoreMetrics(connect *sql.DB, metrics model.KubeScoreRecommendations) {
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO kubescore (id, namespace, cluster_name, recommendations) VALUES (?, ?, ?, ?)")
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

func RetriveKetallEvent(connect *sql.DB) ([]model.Resource, error) {
	rows, err := connect.Query("SELECT resource, namespace, age, cluster_name FROM getall_resources")
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	defer rows.Close()
	var events []model.Resource
	for rows.Next() {
		var result model.Resource
		if err := rows.Scan(&result.Resource, &result.Namespace, &result.Age, &result.ClusterName); err != nil {
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

func RetriveOutdatedEvent(connect *sql.DB) ([]model.CheckResultfinal, error) {
	rows, err := connect.Query("SELECT current_image, current_tag, latest_version, versions_behind, cluster_name FROM outdated_images")
	if err != nil {
		log.Printf("Error: %s", err)
		return nil, err
	}
	defer rows.Close()
	var events []model.CheckResultfinal
	for rows.Next() {
		var result model.CheckResultfinal
		if err := rows.Scan(&result.Image, &result.Current, &result.LatestVersion, &result.VersionsBehind, &result.ClusterName); err != nil {
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

func RetriveKubepugEvent(connect *sql.DB) ([]model.Result, error) {
	rows, err := connect.Query("SELECT result, cluster_name FROM deprecatedAPIs_and_deletedAPIs")
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

func RetrieveEvent(connect *sql.DB) ([]model.DbEvent, error) {
	rows, err := connect.Query("SELECT id, op_type, name, namespace, kind, message, reason, host, event, first_time, last_time, event_time, cluster_name FROM events")
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
