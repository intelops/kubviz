package model

import "github.com/aquasecurity/trivy/pkg/types"

type TrivyImage struct {
	ID          string
	ClusterName string
	Report      types.Report
}
