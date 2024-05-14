package sdk

import (
	"errors"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

type NATSClient struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	config natsConfig
}

func NewNATSClient() (*NATSClient, error) {
	config, err := loadNatsConfig()
	if err != nil {
		return nil, errors.New("Unable to load the nats configurations , error :" + err.Error())
	}
	options := []nats.Option{}
	if config.EnableToken {
		options = append(options, nats.Token(config.NatsToken))
	}
	if config.MtlsConfig.IsEnabled {
		tlsConfig, err := createTLSConfig(config.MtlsConfig)
		if err != nil {
			return nil, err
		}
		options = append(options, nats.Secure(tlsConfig))
	}
	conn, err := nats.Connect(config.NatsAddress, options...)
	if err != nil {
		return nil, err
	}

	js, err := conn.JetStream()
	if err != nil {
		return nil, err
	}

	return &NATSClient{conn: conn, js: js, config: *config}, nil
}

func (natsCli *NATSClient) CreateStream(streamName string) error {
	stream, err := natsCli.js.StreamInfo(streamName)
	log.Printf("Retrieved stream %s", fmt.Sprintf("%v", stream))
	if err != nil {
		log.Printf("Error getting stream %s", err)
	}
	if stream == nil {
		log.Printf("creating stream %q and subjects %q", streamName, streamName+".*")
		_, err = natsCli.js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{streamName + ".*"},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (natsCli *NATSClient) Publish(subject string, data []byte) error {
	_, err := natsCli.js.Publish(subject, data)
	return err
}
