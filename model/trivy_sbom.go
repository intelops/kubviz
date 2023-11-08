package model

import (
	"github.com/aquasecurity/trivy/pkg/sbom/cyclonedx"
)

type Sbom struct {
	ID     string
	Report cyclonedx.BOM
}


