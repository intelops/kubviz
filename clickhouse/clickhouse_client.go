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

func GetClickHouseConnection() (*sql.DB, error) {
	connect, err := sql.Open("clickhouse", "tcp://127.0.0.1:9009?debug=true")
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
			host         String,
			event        String,
			first_time   DateTime,
			last_time    DateTime,
			event_time   DateTime
		) engine=File(TabSeparated)
	`)

	if err != nil {
		log.Fatal(err)
	}
}

func InsertEvent(connect *sql.DB, metrics model.Metrics) {
	var (
		tx, _   = connect.Begin()
		stmt, _ = tx.Prepare("INSERT INTO events (id, op_type, name, namespace, kind, message, host, event, first_time, last_time, event_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	)
	defer stmt.Close()
	eventJson, _ := json.Marshal(metrics.Event)
	if _, err := stmt.Exec(
		metrics.ID,
		metrics.Type,
		metrics.Event.Name,
		metrics.Event.Namespace,
		metrics.Event.InvolvedObject.Kind,
		metrics.Event.Message,
		metrics.Event.Source.Host,
		string(eventJson),
		metrics.Event.FirstTimestamp.Time,
		metrics.Event.LastTimestamp.Time,
		time.Now(),
	); err != nil {
		log.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func RetrieveEvent(connect *sql.DB) {
	rows, err := connect.Query("SELECT country_code, os_id, browser_id, categories, action_day, action_time FROM example")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			country               string
			os, browser           uint8
			categories            []int16
			actionDay, actionTime time.Time
		)
		if err := rows.Scan(&country, &os, &browser, &categories, &actionDay, &actionTime); err != nil {
			log.Fatal(err)
		}
		log.Printf("country: %s, os: %d, browser: %d, categories: %v, action_day: %s, action_time: %s", country, os, browser, categories, actionDay, actionTime)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}
