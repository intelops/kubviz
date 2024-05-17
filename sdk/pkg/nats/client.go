// /pkg/nats/client.go
package nats

import (
	"fmt"
	"log"
	"os"

	"github.com/nats-io/nats.go"
)

type Client struct {
	js     nats.JetStreamContext
	logger *log.Logger
}

func NewClient(cfg *Config) (*Client, error) {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	opts := []nats.Option{nats.Token(cfg.Token)}

	conn, err := nats.Connect(cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("error connecting to NATS: %v", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("error obtaining JetStream context: %v", err)
	}

	return &Client{
		js:     js,
		logger: logger,
	}, nil
}
