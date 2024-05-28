package sdk

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
)

func createTLSConfig(config mtlsConfig) (*tls.Config, error) {
	certPEM, keyPEM, CACertPEM, err := readMtlsCerts(config.CertificateFilePath, config.KeyFilePath, config.CAFilePath)
	if err != nil {
		return nil, errors.New("unable to read the mtls certificates error:" + err.Error())
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("error loading X509 key pair from PEM: %w", err)
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

func readMtlsCerts(certificateFilePath, keyFilePath, CAFilePath string) (certPEM, keyPEM, CACertPEM []byte, err error) {
	certPEM, err = readMtlsFileContents(certificateFilePath)
	if err != nil {
		err = fmt.Errorf("error while reading cert file: %w", err)
		return
	}

	keyPEM, err = readMtlsFileContents(keyFilePath)
	if err != nil {
		err = fmt.Errorf("error while reading key file: %w", err)
		return
	}

	CACertPEM, err = readMtlsFileContents(CAFilePath)
	if err != nil {
		err = fmt.Errorf("error while reading CAcert file: %w", err)
		return
	}

	return

}

func openMtlsCertFile(filepath string) (f *os.File, err error) {
	f, err = os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open mtls certificate file: %w", err)
	}
	return f, nil
}

func readMtlsFileContents(filePath string) ([]byte, error) {
	file, err := openMtlsCertFile(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error while reading file %s:%w", filePath, err)
	}

	return contents, nil
}
