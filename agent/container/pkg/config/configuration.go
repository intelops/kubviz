package config

// var token string = "UfmrJOYwYCCsgQvxvcfJ3BdI6c8WBbnD"
// var natsurl string = "nats://localhost:4222"

// Config will have the configuration details
type Config struct {
	NatsAddress string `envconfig:"NATS_ADDRESS"`
	NatsToken   string `envconfig:"NATS_TOKEN"`
	Port        int    `envconfig:"PORT"`
	StreamName  string `envconfig:"STREAM_NAME"`
}

type GithubConfig struct {
	Org   string `envconfig:"GITHUB_ORG"`
	Token string `envconfig:"GITHUB_TOKEN"`
}
