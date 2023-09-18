package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type AgentConfigurations struct {
	SANamespace            string `envconfig:"SA_NAMESPACE" default:"default"`
	SAName                 string `envconfig:"SA_NAME" default:"default"`
	OutdatedInterval       string `envconfig:"OUTDATED_INTERVAL" default:"@every 5m"`
	GetAllInterval         string `envconfig:"GETALL_INTERVAL" default:"*/30 * * * *"`
	KubeScoreInterval      string `envconfig:"KUBESCORE_INTERVAL" default:"*/40 * * * *"`
	RakkessInterval        string `envconfig:"RAKKESS_INTERVAL" default:"*/50 * * * *"`
	KubePreUpgradeInterval string `envconfig:"KUBEPREUPGRADE_INTERVAL" default:"*/60 * * * *"`
	TrivyInterval          string `envconfig:"TRIVY_INTERVAL" default:"*/10 * * * *"`
	SchedulerEnable        bool   `envconfig:"SCHEDULER_ENABLE" default:"true"`
}

func GetAgentConfigurations() (serviceConf *AgentConfigurations, err error) {
	serviceConf = &AgentConfigurations{}
	if err = envconfig.Process("", serviceConf); err != nil {
		return nil, errors.WithStack(err)
	}
	return
}
