package clients

import (
	"context"
	"fmt"

	"github.com/intelops/kubviz/agent/git/pkg/config"
	"github.com/intelops/kubviz/credential"
	"github.com/intelops/kubviz/model"

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

	cred, err := credential.GetGenericCredential(context.Background(), conf.EntityName, conf.CredIdentifier)
	if err != nil {
		return nil, err
	}
	token:=cred["nats"]

	conn, err := nats.Connect(conf.NatsAddress, nats.Name("Github metrics"), nats.Token(token))
	if err != nil {
		return nil, err
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
	msg := nats.NewMsg(eventSubject)
	msg.Data = metric
	msg.Header.Set("GitProvider", repo)
	msg.Header.Set(string(eventkey), string(eventvalue))
	_, err := n.stream.PublishMsgAsync(msg)

	return err
}
