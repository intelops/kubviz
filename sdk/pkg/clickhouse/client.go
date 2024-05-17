// /pkg/clickhouse/client.go
package clickhouse

import (
	"database/sql"
	"fmt"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type Client struct {
	db *sql.DB
}

func NewClient(cfg *Config) (*Client, error) {
	dataSourceName := fmt.Sprintf("tcp://%s:%d", cfg.DBAddress, cfg.DBPort)

	db, err := sql.Open("clickhouse", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Client{db: db}, nil
}
