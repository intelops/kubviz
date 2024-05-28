package sdk

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type natsConfig struct {
	NatsAddress string `envconfig:"NATS_ADDRESS"`
	NatsToken   string `envconfig:"NATS_TOKEN"`
	MtlsConfig  mtlsConfig
	EnableToken bool `envconfig:"ENABLE_TOKEN"`
}

type mtlsConfig struct {
	CertificateFilePath string `envconfig:"CERT_FILE" default:""`
	KeyFilePath         string `envconfig:"KEY_FILE" default:""`
	CAFilePath          string `envconfig:"CA_FILE" default:""`
	IsEnabled           bool   `envconfig:"ENABLE_MTLS_NATS" default:"false"`
}

func loadNatsConfig() (*natsConfig, error) {
	natsConf := &natsConfig{}
	if err := envconfig.Process("", natsConf); err != nil {
		return nil, errors.WithStack(err)
	}
	return natsConf, nil
}
