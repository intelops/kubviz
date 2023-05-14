package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type AgentConfigurations struct {
	Host                         string `envconfig:"AGENT_HOST" default:"http://localhost"`
	GrpcPort                     string `envconfig:"AGENT_GRPC_PORT" default:":56000"`
	HttpPort                     string `envconfig:"AGENT_HTTP_PORT" default:":8080"`
	SANamespace                  string `envconfig:"SA_NAMESPACE" default:"default"`
	SAName                       string `envconfig:"SA_NAME" default:""`
	TestkubeHost                 string `envconfig:"TESTKUBE_HOST" default:"http://localhost"`
	TestkubePort                 string `envconfig:"TESTKUBE_PORT" default:":30001"`
	PrometheusHost               string `envconfig:"PROMETHEUS_HOST" default:"http://prometheus.default"`
	PrometheusPort               string `envconfig:"PROMETHEUS_PORT" default:":30000"`
	EnvFailureDeterminationValue int    `envconfig:"TEST_FAIL_DETER_PERCENT" default:"80"` //in percentage
	PodStatusCheckRetryTimeout   int    `envconfig:"POD_STATUS_CHECK_TIMEOUT" default:"1"` //in minutes
}

func GetAgentConfigurations() (serviceConf *AgentConfigurations, err error) {
	serviceConf = &AgentConfigurations{}
	if err = envconfig.Process("", serviceConf); err != nil {
		return nil, errors.WithStack(err)
	}
	return
}
