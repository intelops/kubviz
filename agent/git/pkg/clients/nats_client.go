package clients

import (
	"context"
	"fmt"

	"github.com/intelops/kubviz/agent/git/pkg/config"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/mtlsnats"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// constant variables to use with nats stream and
// nats publishing
const (
	StreamName     = "GITMETRICS"
	streamSubjects = "GITMETRICS.*"
	eventSubject   = "GITMETRICS.git"
)

type NATSContext struct {
	conf   *config.Config
	conn   *nats.Conn
	stream nats.JetStreamContext
}

func NewNATSContext(conf *config.Config) (*NATSContext, error) {
	fmt.Println("Waiting before connecting to NATS at:", conf.NatsAddress)
	time.Sleep(1 * time.Second)

	//conn, err := nats.Connect(conf.NatsAddress, nats.Name("Github metrics"), nats.Token(conf.NatsToken))

	var conn *nats.Conn
	var err error
	var mtlsConfig mtlsnats.MtlsConfig

	if mtlsConfig.IsEnabled {
		tlsConfig, err := mtlsnats.GetTlsConfig()
		if err != nil {
			log.Println("Error while getting TLS config:", err)
			return nil, err
		}

		conn, err = nats.Connect(conf.NatsAddress,
			nats.Name("Github metrics"),
			nats.Secure(tlsConfig),
		)
		if err != nil {
			log.Fatal("Error while connecting with mTLS:", err)
		}
	} else {
		conn, err = nats.Connect(conf.NatsAddress, nats.Name("Github metrics"), nats.Token(conf.NatsToken))
		if err != nil {
			log.Println("Error while connecting with token:", err)
			return nil, err
		}
	}

	ctx := &NATSContext{
		conf: conf,
		conn: conn,
	}

	stream, err := ctx.CreateStream()
	if err != nil {
		ctx.conn.Close()
		return nil, err
	}
	ctx.stream = stream

	return ctx, nil
}

func (n *NATSContext) CreateStream() (nats.JetStreamContext, error) {
	// Creates JetStreamContext
	stream, err := n.conn.JetStream()
	if err != nil {
		return nil, err
	}
	// Creates stream
	err = n.checkNAddStream(stream)
	if err != nil {
		return nil, err
	}
	return stream, nil

}

// createStream creates a stream by using JetStreamContext
func (n *NATSContext) checkNAddStream(js nats.JetStreamContext) error {
	// Check if the METRICS stream already exists; if not, create it.
	stream, err := js.StreamInfo(StreamName)
	if err != nil {
		log.Printf("Error getting stream %s", err)
	}
	log.Printf("Retrieved stream %s", fmt.Sprintf("%v", stream))
	if stream == nil {
		log.Printf("creating stream %q and subjects %q", StreamName, streamSubjects)
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     StreamName,
			Subjects: []string{streamSubjects},
		})
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (n *NATSContext) Close() {
	n.conn.Close()
}

func (n *NATSContext) Publish(metric []byte, repo string, eventkey model.EventKey, eventvalue model.EventValue) error {

	ctx := context.Background()
	tracer := otel.Tracer("git-nats-client")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "GitPublish")
	span.SetAttributes(attribute.String("repo-name", repo))
	defer span.End()

	msg := nats.NewMsg(eventSubject)
	msg.Data = metric
	msg.Header.Set("GitProvider", repo)
	msg.Header.Set(string(eventkey), string(eventvalue))
	_, err := n.stream.PublishMsgAsync(msg)

	return err
}
