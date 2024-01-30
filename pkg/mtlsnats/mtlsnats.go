package mtlsnats

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type MtlsConfig struct {
	CertificateFilePath string `envconfig:"CERT_FILE" default:""`
	KeyFilePath         string `envconfig:"KEY_FILE" default:""`
	CAFilePath          string `envconfig:"CA_FILE" default:""`
	IsEnabled           bool   `envconfig:"IS_ENABLED" default:"false"`
}

func ReadMtlsCerts(certificateFilePath, keyFilePath, CAFilePath string) (certPEM, keyPEM, CACertPEM []byte, err error) {
	certPEM, err = ReadMtlsFileContents(certificateFilePath)
	if err != nil {
		err = fmt.Errorf("error while reading cert file: %w", err)
		return
	}

	keyPEM, err = ReadMtlsFileContents(keyFilePath)
	if err != nil {
		err = fmt.Errorf("error while reading key file: %w", err)
		return
	}

	CACertPEM, err = ReadMtlsFileContents(CAFilePath)
	if err != nil {
		err = fmt.Errorf("error while reading CAcert file: %w", err)
		return
	}

	return

}

func OpenMtlsCertFile(filepath string) (f *os.File, err error) {
	f, err = os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open mtls certificate file: %w", err)
	}
	return f, nil
}

func ReadMtlsFileContents(filePath string) ([]byte, error) {
	file, err := OpenMtlsCertFile(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Error while reading file %s:%w", filePath, err)
	}

	return contents, nil
}

func GetTlsConfig() (*tls.Config, error) {

	var cfg MtlsConfig
	err := envconfig.Process("", &cfg)

	if err != nil {
		return nil, fmt.Errorf("Unable to read mtls config %w", err)

	}

	certPEM, keyPEM, CACertPEM, err := ReadMtlsCerts(cfg.CertificateFilePath, cfg.KeyFilePath, cfg.CAFilePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read mtls certificates %w", err)

	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("Error loading X509 key pair from PEM: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(CACertPEM)
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}

	return tlsConfig, nil
}
