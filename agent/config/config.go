package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type AgentConfigurations struct {
	SANamespace      string `envconfig:"SA_NAMESPACE" default:"default"`
	SAName           string `envconfig:"SA_NAME" default:"default"`
	OutdatedInterval string `envconfig:"OUTDATED_INTERVAL" default:"*/20 * * * *"`
}

func GetAgentConfigurations() (serviceConf *AgentConfigurations, err error) {
	serviceConf = &AgentConfigurations{}
	if err = envconfig.Process("", serviceConf); err != nil {
		return nil, errors.WithStack(err)
	}
	return
}
