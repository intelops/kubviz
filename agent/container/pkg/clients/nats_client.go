package clients

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/intelops/kubviz/agent/container/pkg/config"

	"github.com/nats-io/nats.go"
)

// constant variables to use with nats stream and
// nats publishing
const (
	StreamName     = "CONTAINERMETRICS"
	streamSubjects = "CONTAINERMETRICS.*"
	eventSubject   = "CONTAINERMETRICS.git"
)

var (
	certFilePath string = os.Getenv("CERT_FILE")
	keyFilePath  string = os.Getenv("KEY_FILE")
	caFilePath   string = os.Getenv("CA_FILE")
)

// NATSContext encapsulates the connection and context for interacting with a NATS server
// and its associated JetStream. It includes the following fields:
//   - conf: The configuration used to establish the connection, including server address, tokens, etc.
//   - conn: The active connection to the NATS server, allowing for basic NATS operations.
//   - stream: The JetStream context, enabling more advanced stream-based operations such as publishing and subscribing to messages.
//
// NATSContext is used throughout the application to send and receive messages via NATS, and to manage streams within JetStream.
type NATSContext struct {
	conf   *config.Config
	conn   *nats.Conn
	stream nats.JetStreamContext
}

// NewNATSContext establishes a connection to a NATS server using the provided configuration
// and initializes a JetStream context. It checks for the existence of a specific stream
// and creates the stream if it is not found. The function returns a NATSContext object,
// which encapsulates the NATS connection and JetStream context, allowing for publishing
// and subscribing to messages within the application. An error is returned if the connection
// or stream initialization fails.
func NewNATSContext(conf *config.Config) (*NATSContext, error) {
	fmt.Println("Waiting before connecting to NATS at:", conf.NatsAddress)
	time.Sleep(1 * time.Second)
	tlsConfig, err := GetTlsConfig()
	if err != nil {
		log.Println("error while getting tls config ", err)
		time.Sleep(time.Minute * 30)
		log.Fatal("error while getting tls config ", err)
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

// CreateStream initializes a new JetStream within the NATS server, using the configuration
// stored in the NATSContext. It returns the JetStream context, allowing for further interaction
// with the stream, such as publishing and subscribing to messages. If the stream creation fails,
// an error is returned. This method is typically called during initialization to ensure that
// the required stream is available for the application's messaging needs.
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

// Publish sends a given event to the JetStream within the NATS server, including the repository information in the header.
// The event is provided as a byte slice, and the target repository information is identified by the repo string.
// This method leverages the JetStream context within the NATSContext to publish the event, ensuring that it is sent with the correct headers and to the appropriate stream within the NATS server.
// The repository information in the header can be used by subscribers to filter or route the event based on its origin or destination.
// An error is returned if the publishing process fails, such as if the connection is lost or if there are issues with the JetStream.
func (n *NATSContext) Publish(event []byte, repo string) error {
	msg := nats.NewMsg(eventSubject)
	msg.Data = event
	msg.Header.Set("REPO_NAME", repo)
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

func GetTlsConfig() (*tls.Config, error) {
	certPEM, keyPEM, caCertPEM, err := ReadMtlsCerts(certFilePath, keyFilePath, caFilePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read mtls certs %w", err)
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("Error loading X509 key pair from PEM: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertPEM)
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}
	return tlsConfig, nil
}
