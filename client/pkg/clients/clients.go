package clients

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/nats-io/nats.go"
)

var (
	certFilePath string = os.Getenv("CLIENT_CERT_FILE")
	keyFilePath  string = os.Getenv("CLIENT_KEY_FILE")
	caFilePath   string = os.Getenv("CLIENT_CA_FILE")
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
