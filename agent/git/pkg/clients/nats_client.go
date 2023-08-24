package clients

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"

	"github.com/intelops/kubviz/agent/git/pkg/config"
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

var (
	certFilePath string = os.Getenv("CERT_FILE")
	keyFilePath  string = os.Getenv("KEY_FILE")
	caFilePath   string = os.Getenv("CA_FILE")
)

type NATSContext struct {
	conf   *config.Config
	conn   *nats.Conn
	stream nats.JetStreamContext
}

func NewNATSContext(conf *config.Config) (*NATSContext, error) {
	fmt.Println("Waiting before connecting to NATS at:", conf.NatsAddress)
	time.Sleep(1 * time.Second)
	certPEM, keyPEM, caCertPEM, err := ReadMtlsCerts(certFilePath, keyFilePath, caFilePath)
	if err != nil {
		log.Printf("Error reading MTLS certs: %v", err)
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Printf("Error loading X509 key pair from PEM: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertPEM)
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}
	conn, err := nats.Connect(conf.NatsAddress,
		nats.Name("Github metrics"),
		nats.Token(conf.NatsToken),
		nats.Secure(tlsConfig),
	)
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

func ReadMtlsCerts(certFile, keyFile, caFile string) (certPEM, keyPEM, caCertPEM []byte, err error) {
	cf, err := OpenMtlsCertFile(certFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error opening cert file: %w", err)
	}
	defer cf.Close()

	certPEM, err = io.ReadAll(cf)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error reading cert file: %w", err)
	}

	kf, err := OpenMtlsCertFile(keyFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error opening key file: %w", err)
	}
	defer kf.Close()

	keyPEM, err = io.ReadAll(kf)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error reading key file: %w", err)
	}

	caf, err := OpenMtlsCertFile(caFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error opening CA file: %w", err)
	}
	defer caf.Close()

	caCertPEM, err = io.ReadAll(caf)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error reading CA file: %w", err)
	}

	return certPEM, keyPEM, caCertPEM, nil
}

func OpenMtlsCertFile(path string) (f *os.File, err error) {
	f, err = os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open MTLS cert file: %w", err)
	}
	return f, nil
}
