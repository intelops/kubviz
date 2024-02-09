package config

type Config struct {
	NatsAddress        string `envconfig:"NATS_ADDRESS"`
	NatsToken          string `envconfig:"NATS_TOKEN"`
	DbPort             int    `envconfig:"DB_PORT"`
	DBAddress          string `envconfig:"DB_ADDRESS"`
	ClickHouseUsername string `envconfig:"CLICKHOUSE_USERNAME"`
	ClickHousePassword string `envconfig:"CLICKHOUSE_PASSWORD"`
	KetallConsumer     string `envconfig:"KETALL_EVENTS_CONSUMER"`
	RakeesConsumer     string `envconfig:"RAKEES_METRICS_CONSUMER"`
	OutdatedConsumer   string `envconfig:"OUTDATED_EVENTS_CONSUMER"`
	DeprecatedConsumer string `envconfig:"DEPRECATED_API_CONSUMER"`
	DeletedConsumer    string `envconfig:"DELETED_API_CONSUMER"`
	KubvizConsumer     string `envconfig:"KUBVIZ_EVENTS_CONSUMER"`
	KubscoreConsumer   string `envconfig:"KUBSCORE_CONSUMER"`
	TrivyConsumer      string `envconfig:"TRIVY_CONSUMER"`
	TrivyImageConsumer string `envconfig:"TRIVY_IMAGE_CONSUMER"`
	TrivySbomConsumer  string `envconfig:"TRIVY_SBOM_CONSUMER"`
}
