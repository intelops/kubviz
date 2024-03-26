package model

type Sbom struct {
	ID          string
	ClusterName string
	Report      map[string]interface{}
}
