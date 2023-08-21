package model

import (
	"time"
)

type Reports struct {
	ID     string
	Report Sbom
}

type Sbom struct {
	Schema       string `json:"$schema"`
	BomFormat    string `json:"bomFormat"`
	SpecVersion  string `json:"specVersion"`
	SerialNumber string `json:"serialNumber"`
	Version      int    `json:"version"`
	Metadata     struct {
		Timestamp time.Time `json:"timestamp"`
		Tools     []struct {
			Vendor  string `json:"vendor"`
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"tools"`
		Component struct {
			BomRef     string `json:"bom-ref"`
			Type       string `json:"type"`
			Name       string `json:"name"`
			Purl       string `json:"purl"`
			Properties []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"properties"`
		} `json:"component"`
	} `json:"metadata"`
	Components []struct {
		BomRef     string `json:"bom-ref"`
		Type       string `json:"type"`
		Name       string `json:"name"`
		Version    string `json:"version"`
		Properties []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"properties"`
		Hashes []struct {
			Alg     string `json:"alg"`
			Content string `json:"content"`
		} `json:"hashes,omitempty"`
		Licenses []struct {
			Expression string `json:"expression"`
		} `json:"licenses,omitempty"`
		Purl string `json:"purl,omitempty"`
	} `json:"components"`
	Dependencies []struct {
		Ref       string   `json:"ref"`
		DependsOn []string `json:"dependsOn"`
	} `json:"dependencies"`
	Vulnerabilities []interface{} `json:"vulnerabilities"`
}
