package clients

import (
	"errors"
	"log"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/nats-io/nats.go"
)

// ErrHeaderEmpty defines an error occur when header is empty in git stream
var (
	ErrHeaderEmpty = errors.New("git header is empty while subscribing from agent")
)

// GitNats specifies a Git related jetstream subjects, subject and consumer names
type GitNats string

const (
	bridgeSubjects GitNats = "GITMETRICS.*"
	bridgeSubject  GitNats = "GITMETRICS.git"
	bridgeConsumer GitNats = "Git-Consumer"
)

// SubscribeGitBridgeNats subscribes to nats jetstream and calls
// the respective funcs to insert data into clickhouse DB
func (n *NATSContext) SubscribeGitBridgeNats(conn clickhouse.DBInterface) {
	n.stream.Subscribe(string(bridgeSubject), func(msg *nats.Msg) {
		// Recover from a panic
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from panic:", r)

				// Acknowledge the message
				msg.Ack()
			}
		}()
		msg.Ack()

	}, nats.Durable(string(bridgeConsumer)), nats.ManualAck())
}
