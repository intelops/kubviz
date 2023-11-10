package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DbPort             int    `envconfig:"DB_PORT" required:"true"`
	DBAddress          string `envconfig:"DB_ADDRESS" required:"true"`
	ClickHouseUsername string `envconfig:"CLICKHOUSE_USERNAME"`
	ClickHousePassword string `envconfig:"CLICKHOUSE_PASSWORD"`
	SchemaPath         string `envconfig:"SCHEMA_PATH" default:"/sql"`
	TtlInterval        string `envconfig:"TTL_INTERVAL" default:"1"`
	TtlUnit            string `envconfig:"TTL_UNIT" default:"MONTH"`
}

func New() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
