package model

import "github.com/aquasecurity/trivy/pkg/k8s/report"

type Trivy struct {
	ID          string
	ClusterName string
	Report      report.ConsolidatedReport
}
