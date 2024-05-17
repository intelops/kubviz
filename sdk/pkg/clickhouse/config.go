package clickhouse

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBAddress string `envconfig:"DB_ADDRESS" default:"localhost"`
	DBPort    int    `envconfig:"DB_PORT" default:"9000"`
	Username  string `envconfig:"CLICKHOUSE_USERNAME"`
	Password  string `envconfig:"CLICKHOUSE_PASSWORD"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
