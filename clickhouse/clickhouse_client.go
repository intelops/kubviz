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
        CREATE TABLE IF NOT EXISTS DeprecatedAPIs (
			ClusterName String,
            Description String,
            Kind String,
            Deprecated UInt8,
            Scope String,
            ObjectName String
        ) engine=File(TabSeparated)
    `)
	if err != nil {
		log.Fatal(err)
	}

	_, err = connect.Exec(`
        CREATE TABLE IF NOT EXISTS DeletedAPIs (
			ClusterName String,
            Group String,
            Kind String,
            Version String,
            Name String,
            Deleted UInt8,
            Scope String,
            ObjectName String
        ) engine=File(TabSeparated)
    `)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateKetallSchema(connect *sql.DB) {
	_, err := connect.Exec(`
		CREATE TABLE IF NOT EXISTS getall_resources (
			Cluster_Name String,
			Namespace String,
			Kind String,
			Resource String,
			Age String
        ) engine=File(TabSeparated)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func CreateOutdatedSchema(connect *sql.DB) {
	_, err := connect.Exec(`
	    CREATE TABLE IF NOT EXISTS outdated_images (
			Cluster_Name String,
			Namespace String,
			Pod String,
		    Current_Image String,
			Current_Tag String,
			Latest_Version String,
			Versions_Behind Int64
	    ) engine=File(TabSeparated)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func InsertKetallEvent(connect *sql.DB, metrics model.Resource) {
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO getall_resources (Cluster_Name, Namespace, Kind, Resource, Age) VALUES (?, ?, ?, ?, ?)")
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

func InsertOutdatedEvent(connect *sql.DB, metrics model.CheckResultfinal) {
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO outdated_images (Cluster_Name, Namespace, Pod, Current_Image, Current_Tag, Latest_Version, Versions_Behind) VALUES (?, ?, ?, ?, ?, ?, ?)")
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

func InsertDeprecatedAPI(connect *sql.DB, deprecatedAPI model.DeprecatedAPI) {
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO DeprecatedAPIs (ClusterName, Description, Kind, Deprecated, Scope, ObjectName) VALUES (?, ?, ?, ?, ?, ?)")
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

func InsertDeletedAPI(connect *sql.DB, deletedAPI model.DeletedAPI) {
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO DeletedAPIs (ClusterName, Group, Kind, Version, Name, Deleted, Scope, ObjectName) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
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

func RetriveKetallEvent(connect *sql.DB) ([]model.Resource, error) {
	rows, err := connect.Query("SELECT Cluster_Name, Namespace, Kind, Resource, Age  FROM getall_resources")
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

func RetriveOutdatedEvent(connect *sql.DB) ([]model.CheckResultfinal, error) {
	rows, err := connect.Query("SELECT Cluster_name, Namespace, Pod, Current_Image, Current_Tag, Latest_Version, Versions_Behind FROM outdated_images")
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
