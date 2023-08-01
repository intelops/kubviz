package config

type Config struct {
	NatsAddress string `envconfig:"NATS_ADDRESS"`
	CredIdentifier string `envconfig:"NATS_CRED_IDENTIFIER" default:"authToken"`
	EntityName     string `envconfig:"NATS_ENTITY_NAME" default:"astra"`

	//NatsToken   string `envconfig:"NATS_TOKEN"`
	DbPort      int    `envconfig:"DB_PORT"`
	DBAddress   string `envconfig:"DB_ADDRESS"`
}
