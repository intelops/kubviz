package scheduler

import (
	"github.com/intelops/kubviz/agent/kubviz/plugins/events"
	"github.com/intelops/kubviz/agent/kubviz/plugins/ketall"
	"github.com/intelops/kubviz/agent/kubviz/plugins/kubepreupgrade"
	"github.com/nats-io/nats.go"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/intelops/kubviz/agent/kubviz/plugins/kubescore"
	"github.com/intelops/kubviz/agent/kubviz/plugins/outdated"
	"github.com/intelops/kubviz/agent/kubviz/plugins/rakkess"
	"github.com/intelops/kubviz/agent/kubviz/plugins/trivy"
)

type OutDatedImagesJob struct {
	config    *rest.Config
	js        nats.JetStreamContext
	frequency string
}

type KetallJob struct {
	config    *rest.Config
	js        nats.JetStreamContext
	frequency string
}
type TrivyJob struct {
	config    *rest.Config
	js        nats.JetStreamContext
	frequency string
}
type RakkessJob struct {
	config    *rest.Config
	js        nats.JetStreamContext
	frequency string
}
type KubePreUpgradeJob struct {
	config    *rest.Config
	js        nats.JetStreamContext
	frequency string
}
type KubescoreJob struct {
	clientset *kubernetes.Clientset
	js        nats.JetStreamContext
	frequency string
}

func NewOutDatedImagesJob(config *rest.Config, js nats.JetStreamContext, frequency string) (*OutDatedImagesJob, error) {
	return &OutDatedImagesJob{
		config:    config,
		js:        js,
		frequency: frequency,
	}, nil
}
func (v *OutDatedImagesJob) CronSpec() string {
	return v.frequency
}

func (j *OutDatedImagesJob) Run() {
	// Call the outDatedImages function with the provided config and js
	err := outdated.OutDatedImages(j.config, j.js)
	events.LogErr(err)
}
func NewKetallJob(config *rest.Config, js nats.JetStreamContext, frequency string) (*KetallJob, error) {
	return &KetallJob{
		config:    config,
		js:        js,
		frequency: frequency,
	}, nil
}
func (v *KetallJob) CronSpec() string {
	return v.frequency
}

func (j *KetallJob) Run() {
	// Call the Ketall function with the provided config and js
	err := ketall.GetAllResources(j.config, j.js)
	events.LogErr(err)
}

func NewKubePreUpgradeJob(config *rest.Config, js nats.JetStreamContext, frequency string) (*KubePreUpgradeJob, error) {
	return &KubePreUpgradeJob{
		config:    config,
		js:        js,
		frequency: frequency,
	}, nil
}
func (v *KubePreUpgradeJob) CronSpec() string {
	return v.frequency
}

func (j *KubePreUpgradeJob) Run() {
	// Call the Kubepreupgrade function with the provided config and js
	err := kubepreupgrade.KubePreUpgradeDetector(j.config, j.js)
	events.LogErr(err)
}

func NewKubescoreJob(clientset *kubernetes.Clientset, js nats.JetStreamContext, frequency string) (*KubescoreJob, error) {
	return &KubescoreJob{
		clientset: clientset,
		js:        js,
		frequency: frequency,
	}, nil
}
func (v *KubescoreJob) CronSpec() string {
	return v.frequency
}

func (j *KubescoreJob) Run() {
	// Call the Kubescore function with the provided config and js
	err := kubescore.RunKubeScore(j.clientset, j.js)
	events.LogErr(err)
}
func NewRakkessJob(config *rest.Config, js nats.JetStreamContext, frequency string) (*RakkessJob, error) {
	return &RakkessJob{
		config:    config,
		js:        js,
		frequency: frequency,
	}, nil
}
func (v *RakkessJob) CronSpec() string {
	return v.frequency
}

func (j *RakkessJob) Run() {
	// Call the Rakkes function with the provided config and js
	err := rakkess.RakeesOutput(j.config, j.js)
	events.LogErr(err)
}
func NewTrivyJob(config *rest.Config, js nats.JetStreamContext, frequency string) (*TrivyJob, error) {
	return &TrivyJob{
		config:    config,
		js:        js,
		frequency: frequency,
	}, nil
}
func (v *TrivyJob) CronSpec() string {
	return v.frequency
}

func (j *TrivyJob) Run() {
	// Call the Trivy function with the provided config and js
	err := trivy.RunTrivySbomScan(j.config, j.js)
	events.LogErr(err)
	err = trivy.RunTrivyImageScans(j.config, j.js)
	events.LogErr(err)
	err = trivy.RunTrivyK8sClusterScan(j.js)
	events.LogErr(err)
}
