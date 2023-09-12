package main

import (
	"github.com/nats-io/nats.go"
	"k8s.io/client-go/rest"
)

type OutDatedImagesJob struct {
	config    *rest.Config
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
	err := OutDatedImages(j.config, j.js)
	LogErr(err)
}
