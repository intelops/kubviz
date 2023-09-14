package config

type Config struct {
	Enabled        bool   `envconfig:"ENABLED" default:"true"`
	CredIdentifier string `envconfig:"NATS_CRED_IDENTIFIER" default:"authToken"`
	EntityName     string `envconfig:"NATS_ENTITY_NAME" default:"astra"`
	NatsAddress    string `envconfig:"NATS_ADDRESS"`
	NatsToken      string `envconfig:"NATS_TOKEN"`
	DbPort         int    `envconfig:"DB_PORT"`
	DBAddress      string `envconfig:"DB_ADDRESS"`
}
