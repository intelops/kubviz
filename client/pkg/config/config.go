package config

type Config struct {
	NatsAddress        string `envconfig:"NATS_ADDRESS"`
	NatsToken          string `envconfig:"NATS_TOKEN"`
	DbPort             int    `envconfig:"DB_PORT"`
	DBAddress          string `envconfig:"DB_ADDRESS"`
	ClickHouseUsername string `envconfig:"CLICKHOUSE_USERNAME"`
	ClickHousePassword string `envconfig:"CLICKHOUSE_PASSWORD"`
	AWSRegion          string `envconfig:"AWS_REGION" default:""`
	AWSAccessKey       string `envconfig:"AWS_ACCESS_KEY" default:""`
	AWSSecretKey       string `envconfig:"AWS_SECRET_KEY" default:""`
	S3BucketName       string `envconfig:"S3_BUCKET_NAME" default:""`
	S3ObjectKey        string `envconfig:"S3_OBJECT_KEY" default:""`
}
