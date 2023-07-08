package clients

import (
	"fmt"
	"log"
	"time"

	"github.com/kube-tarian/kubviz/client/pkg/clickhouse"
	"github.com/kube-tarian/kubviz/client/pkg/config"
	"github.com/nats-io/nats.go"
)

type NATSContext struct {
	conf     *config.Config
	conn     *nats.Conn
	stream   nats.JetStreamContext
	dbClient clickhouse.DBInterface
}

func NewNATSContext(conf *config.Config, dbClient clickhouse.DBInterface) (*NATSContext, error) {
	log.Println("Waiting before connecting to NATS at:", conf.NatsAddress)
	time.Sleep(1 * time.Second)

	conn, err := nats.Connect(conf.NatsAddress, nats.Name("Github metrics"), nats.Token(conf.NatsToken))
	if err != nil {
		return nil, err
	}

	ctx := &NATSContext{
		conf:     conf,
		conn:     conn,
		dbClient: dbClient,
	}

	stream, err := ctx.createStream()
	if err != nil {
		ctx.conn.Close()
		return nil, err
	}

	ctx.stream = stream

	_, err = stream.StreamInfo("GITMETRICS")
	if err != nil {
		return nil, fmt.Errorf("git metrics stream not found %w", err)
	}
	ctx.SubscribeGitBridgeNats(dbClient)

	_, err = stream.StreamInfo("CONTAINERMETRICS")
	if err != nil {
		return nil, fmt.Errorf("container metrics stream not found %w", err)
	}
	ctx.SubscribeContainerNats(dbClient)
	_, err = stream.StreamInfo("METRICS")
	if err != nil {
		return nil, fmt.Errorf("kubeviz metrics stream not found %w", err)
	}
	ctx.SubscribeAllKubvizNats(dbClient)

	return ctx, nil
}
func (n *NATSContext) createStream() (nats.JetStreamContext, error) {
	// Creates JetStreamContext
	stream, err := n.conn.JetStream()
	if err != nil {
		return nil, err
	}
	return stream, nil
}
func (n *NATSContext) Close() {
	n.conn.Close()
	n.dbClient.Close()
}
