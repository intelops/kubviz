package model

import v1 "k8s.io/api/core/v1"

type Metrics struct {
	ID          string
	Type        string
	Event       *v1.Event
	ClusterName string
}
