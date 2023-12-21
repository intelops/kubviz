package main

import (
	"github.com/nats-io/nats.go"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	err := outDatedImages(j.config, j.js)
	LogErr(err)
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
	err := GetAllResources(j.config, j.js)
	LogErr(err)
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
	err := KubePreUpgradeDetector(j.config, j.js)
	LogErr(err)
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
	err := RunKubeScore(j.clientset, j.js)
	LogErr(err)
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
	err := RakeesOutput(j.config, j.js)
	LogErr(err)
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
	err := RunTrivySbomScan(j.config, j.js)
	LogErr(err)
	// err := runTrivyScans(j.config, j.js)

}
