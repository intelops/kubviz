package config

import (
	"database/sql"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DbPort     int    `envconfig:"DB_PORT" required:"true"`
	DBAddress  string `envconfig:"DB_ADDRESS" required:"true"`
	SchemaPath string `envconfig:"SCHEMA_PATH" default:"/sql"`
}

func OpenClickHouseConn() (*sql.DB, *Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, nil, err
	}
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.DBAddress, cfg.DbPort)},
	})
	if err := conn.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			return nil, nil, fmt.Errorf("[%d] %s %s", exception.Code, exception.Message, exception.StackTrace)
		} else {
			return nil, nil, err
		}
	}
	return conn, &cfg, nil
}
