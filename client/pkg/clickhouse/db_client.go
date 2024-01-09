package clickhouse

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/intelops/kubviz/gitmodels/dbstatement"
	"github.com/intelops/kubviz/model"
)

type DBClient struct {
	splconn driver.Conn
	conn    *sql.DB
	conf    *config.Config
}
type DBInterface interface {
	InsertRakeesMetrics(model.RakeesMetrics)
	InsertKetallEvent(model.Resource)
	InsertOutdatedEvent(model.CheckResultfinal)
	InsertDeprecatedAPI(model.DeprecatedAPI)
	InsertDeletedAPI(model.DeletedAPI)
	InsertKubvizEvent(model.Metrics)
	InsertGitEvent(string)
	InsertKubeScoreMetrics(model.KubeScoreRecommendations)
	InsertTrivyImageMetrics(metrics model.TrivyImage)
	InsertTrivySbomMetrics(metrics model.SbomData)
	InsertTrivyMetrics(metrics model.Trivy)
	RetriveKetallEvent() ([]model.Resource, error)
	RetriveOutdatedEvent() ([]model.CheckResultfinal, error)
	RetriveKubepugEvent() ([]model.Result, error)
	RetrieveKubvizEvent() ([]model.DbEvent, error)
	InsertContainerEventDockerHub(model.DockerHubBuild)
	InsertContainerEventAzure(model.AzureContainerPushEventPayload)
	InsertContainerEventQuay(model.QuayImagePushPayload)
	InsertContainerEventJfrog(model.JfrogContainerPushEventPayload)
	InsertContainerEventGithub(string)
	InsertGitCommon(metrics model.GitCommonAttribute, statement dbstatement.DBStatement) error
	Close()
}

func NewDBClient(conf *config.Config) (DBInterface, error) {
	splConn, stdConn, err := connect(conf)
	if err != nil {
		return nil, err
	}

	return &DBClient{splconn: splConn, conn: stdConn, conf: conf}, nil
}

func (c *DBClient) Close() {
	_ = c.conn.Close()
}

func DbUrl(conf *config.Config) string {
	return fmt.Sprintf("tcp://%s:%d?debug=true", conf.DBAddress, conf.DbPort)
}

func connect(conf *config.Config) (driver.Conn, *sql.DB, error) {
	ctx := context.Background()
	var connOptions clickhouse.Options

	connOptions.Addr = []string{fmt.Sprintf("%s:%d", conf.DBAddress, conf.DbPort)}
	connOptions.Debug = true

	if conf.ClickHouseUsername != "" && conf.ClickHousePassword != "" {
		connOptions.Auth = clickhouse.Auth{
			Username: conf.ClickHouseUsername,
			Password: conf.ClickHousePassword,
		}
	}

	if conf.EnableTLS {
		cert, err := tls.LoadX509KeyPair(conf.TLSCertPath, conf.TLSKeyPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load TLS certificate: %w", err)
		}

		caCert, err := os.ReadFile(conf.CACertPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}
		connOptions.TLS = tlsConfig
	}
	stdConnOption := connOptions
	stdConn := clickhouse.OpenDB(&stdConnOption)
	if err := stdConn.Ping(); err != nil {
		return nil, nil, err
	}
	connOptions.Settings = clickhouse.Settings{
		"allow_experimental_object_type": 1,
	}
	splConn, err := clickhouse.Open(&connOptions)
	if err != nil {
		return nil, nil, err
	}

	if err := splConn.Ping(ctx); err != nil {
		return nil, nil, err
	}

	return splConn, stdConn, nil
}
