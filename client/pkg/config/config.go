package config

type Config struct {
	NatsAddress        string `envconfig:"NATS_ADDRESS"`
	NatsToken          string `envconfig:"NATS_TOKEN"`
	DbPort             int    `envconfig:"DB_PORT"`
	DBAddress          string `envconfig:"DB_ADDRESS"`
	ClickHouseUsername string `envconfig:"CLICKHOUSE_USERNAME"`
	ClickHousePassword string `envconfig:"CLICKHOUSE_PASSWORD"`
	EnableTLS          bool   `envconfig:"ENABLE_TLS" default:"false"`
	TLSCertPath        string `envconfig:"TLS_CERT_PATH"`
	TLSKeyPath         string `envconfig:"TLS_KEY_PATH"`
	CACertPath         string `envconfig:"CA_CERT_PATH"`
}
