package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/aquasecurity/trivy/pkg/sbom/cyclonedx"
	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	"k8s.io/client-go/rest"
)

func publishTrivySbomReport(report cyclonedx.BOM, js nats.JetStreamContext) error {
	metrics := model.Sbom{
		ID:     uuid.New().String(),
		Report: report,
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.TRIVY_SBOM_SUBJECT, metricsJson)
	if err != nil {
		return err
	}
	log.Printf("Trivy sbom report with Id %v has been published\n", metrics.ID)
	return nil
}

func executeCommandSbom(command string) ([]byte, error) {
	cmd := exec.Command("/bin/sh", "-c", command)
	var outc, errc bytes.Buffer
	cmd.Stdout = &outc
	cmd.Stderr = &errc
	err := cmd.Run()
	if err != nil {
		log.Println("Execute SBOM Command Error", err.Error())
	}
	return outc.Bytes(), err
}

func RunTrivySbomScan(config *rest.Config, js nats.JetStreamContext) error {
	log.Println("trivy sbom scan started...")
	pvcMountPath := "/mnt/agent/kbz"
	trivySbomCacheDir := fmt.Sprintf("%s/trivy-sbomcache", pvcMountPath)
	err := os.MkdirAll(trivySbomCacheDir, 0755)
	if err != nil {
		log.Printf("Error creating Trivy cache directory: %v\n", err)
		return err
	}
	images, err := ListImages(config)
	if err != nil {
		log.Printf("failed to list images: %v", err)
	}
	for _, image := range images {
		sbomcmd := fmt.Sprintf("trivy image --format cyclonedx %s --cache-dir %s", image.PullableImage, trivySbomCacheDir)
		out, err := executeCommandSbom(sbomcmd)
		if err != nil {
			log.Printf("Error executing Trivy for image sbom %s: %v", image.PullableImage, err)
			continue
		}
		if len(out) == 0 {
			log.Printf("Trivy output is empty for image sbom %s", image.PullableImage)
			continue
		}

		var report cyclonedx.BOM
		err = json.Unmarshal(out, &report)
		if err != nil {
			log.Printf("Error unmarshaling JSON data for image sbom %s: %v", image.PullableImage, err)
			continue
		}

		/* if _, err := stmt.Exec(
			metrics.ID,
			result.CycloneDX.Metadata.Component.Name,
			result.CycloneDX.Metadata.Component.PackageURL,
			result.CycloneDX.Metadata.Component.BOMRef,
			result.CycloneDX.SerialNumber,
			int32(result.CycloneDX.Version),
			result.CycloneDX.BOMFormat,
			result.CycloneDX.Metadata.Component.Version,
			result.CycloneDX.Metadata.Component.MIMEType,
		);
		*/
		log.Println("sbom log from agent side:")
		log.Println("sbom log from client side:")
		log.Println("component name :", report.CycloneDX.Metadata.Component.Name)
		log.Println("package url :", report.CycloneDX.Metadata.Component.PackageURL)
		log.Println("bom ref :", report.CycloneDX.Metadata.Component.BOMRef)
		log.Println("serial number :", report.CycloneDX.SerialNumber)
		log.Println("cyclone dx version :", report.CycloneDX.Version)
		log.Println("bom format :", report.CycloneDX.BOMFormat)
		log.Println("component version :", report.CycloneDX.Metadata.Component.Version)
		log.Println("mime type :", report.CycloneDX.Metadata.Component.MIMEType)
		publishTrivySbomReport(report, js)
	}
	return nil
}
