package nats

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Address string `envconfig:"NATS_ADDRESS" default:"nats://localhost:4222"`
	Token   string `envconfig:"NATS_TOKEN"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
