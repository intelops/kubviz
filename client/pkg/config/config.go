package config

type Config struct {
	NatsAddress        string `envconfig:"NATS_ADDRESS"`
	NatsToken          string `envconfig:"NATS_TOKEN"`
	DbPort             int    `envconfig:"DB_PORT"`
	DBAddress          string `envconfig:"DB_ADDRESS"`
	ClickHouseUsername string `envconfig:"CLICKHOUSE_USERNAME"`
	ClickHousePassword string `envconfig:"CLICKHOUSE_PASSWORD"`
	KetallConsumer     string `envconfig:"KETALL_EVENTS_CONSUMER" required:"true"`
	RakeesConsumer     string `envconfig:"RAKEES_METRICS_CONSUMER" required:"true"`
	OutdatedConsumer   string `envconfig:"OUTDATED_EVENTS_CONSUMER" required:"true"`
	DeprecatedConsumer string `envconfig:"DEPRECATED_API_CONSUMER" required:"true"`
	DeletedConsumer    string `envconfig:"DELETED_API_CONSUMER" required:"true"`
	KubvizConsumer     string `envconfig:"KUBVIZ_EVENTS_CONSUMER" required:"true"`
	KubscoreConsumer   string `envconfig:"KUBSCORE_CONSUMER" required:"true"`
	TrivyConsumer      string `envconfig:"TRIVY_CONSUMER" required:"true"`
	TrivyImageConsumer string `envconfig:"TRIVY_IMAGE_CONSUMER" required:"true"`
	TrivySbomConsumer  string `envconfig:"TRIVY_SBOM_CONSUMER" required:"true"`
	AwsEnable          bool   `envconfig:"AWS_ENABLE" default:"false"`
	AWSRegion          string `envconfig:"AWS_REGION"`
	AWSAccessKey       string `envconfig:"AWS_ACCESS_KEY"`
	AWSSecretKey       string `envconfig:"AWS_SECRET_KEY"`
	S3BucketName       string `envconfig:"S3_BUCKET_NAME"`
	S3ObjectKey        string `envconfig:"S3_OBJECT_KEY"`
}
