package sdk

import (
	"log"
	"os"

	"github.com/intelops/kubviz/sdk/pkg/clickhouse"
	"github.com/intelops/kubviz/sdk/pkg/nats"
)

type SDK struct {
	natsClient       *nats.Client
	clickhouseClient *clickhouse.Client
	logger           *log.Logger
}

func New(natsCfg *nats.Config, chCfg *clickhouse.Config) (*SDK, error) {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	natsClient, err := nats.NewClient(natsCfg)
	if err != nil {
		return nil, err
	}

	chClient, err := clickhouse.NewClient(chCfg)
	if err != nil {
		return nil, err
	}

	return &SDK{
		natsClient:       natsClient,
		clickhouseClient: chClient,
		logger:           logger,
	}, nil
}

func (sdk *SDK) Start() error {
	return nil
}
