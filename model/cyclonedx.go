package model

import "github.com/aquasecurity/trivy/pkg/sbom/cyclonedx/core"

type Cyclone struct {
	ID  string
	Rep core.Component
}
