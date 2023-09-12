package config

// var token string = "UfmrJOYwYCCsgQvxvcfJ3BdI6c8WBbnD"
// var natsurl string = "nats://localhost:4222"

// Config will have the configuration details
type Config struct {
	Enabled        bool   `envconfig:"ENABLED"`
	CredIdentifier string `envconfig:"NATS_CRED_IDENTIFIER" default:"authToken"`
	EntityName     string `envconfig:"NATS_ENTITY_NAME" default:"nats"`
	NatsAddress    string `envconfig:"NATS_ADDRESS"`
	NatsToken      string `envconfig:"NATS_"`
	Port           int    `envconfig:"PORT"`
	StreamName     string `envconfig:"STREAM_NAME"`
}

type GithubConfig struct {
	Org   string `envconfig:"GITHUB_ORG"`
	Token string `envconfig:"GITHUB_TOKEN"`
}
