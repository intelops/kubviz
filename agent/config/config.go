package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

type AgentConfigurations struct {
	SANamespace              string `envconfig:"SA_NAMESPACE" default:"default"`
	SAName                   string `envconfig:"SA_NAME" default:"default"`
	OutdatedInterval         string `envconfig:"OUTDATED_INTERVAL" default:"0"`
	GetAllInterval           string `envconfig:"GETALL_INTERVAL" default:"*/30 * * * *"`
	KubeScoreInterval        string `envconfig:"KUBESCORE_INTERVAL" default:"*/40 * * * *"`
	RakkessInterval          string `envconfig:"RAKKESS_INTERVAL" default:"*/50 * * * *"`
	KubePreUpgradeInterval   string `envconfig:"KUBEPREUPGRADE_INTERVAL" default:"*/60 * * * *"`
	TrivyImageInterval       string `envconfig:"TRIVY_IMAGE_INTERVAL" default:"*/10 * * * *"`
	TrivySbomInterval        string `envconfig:"TRIVY_SBOM_INTERVAL" default:"*/20 * * * *"`
	TrivyClusterScanInterval string `envconfig:"TRIVY_CLUSTERSCAN_INTERVAL" default:"*/35 * * * *"`
	SchedulerEnable          bool   `envconfig:"SCHEDULER_ENABLE" default:"true"`
	KuberHealthyEnable       bool   `envconfig:"KUBERHEALTHY_ENABLE" default:"true"`
}

func GetAgentConfigurations() (serviceConf *AgentConfigurations, err error) {
	serviceConf = &AgentConfigurations{}
	if err = envconfig.Process("", serviceConf); err != nil {
		return nil, errors.WithStack(err)
	}
	return
}

type KHConfig struct {
	KuberhealthyURL string        `envconfig:"KUBERHEALTHY_URL" required:"true" default:"test.com"`
	PollInterval    time.Duration `envconfig:"POLL_INTERVAL" default:"0.01s"`
}

func GetKuberHealthyConfig() (khconfig *KHConfig, err error) {
	khconfig = &KHConfig{}
	if err = envconfig.Process("", khconfig); err != nil {
		return nil, errors.WithStack(err)
	}
	return
}
