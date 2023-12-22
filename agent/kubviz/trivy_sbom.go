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

	for _,packageinfo :=range report.Packages {
		for _, pkg := range packageinfo.Packages {
			//log.Println(pkg.Name)


	metrics := model.SbomData{
		ID:               uuid.New().String(),
		ClusterName: ClusterName,
		ComponentName:    report.CycloneDX.Metadata.Component.Name,
		PackageName: pkg.Name,
		PackageUrl:       report.CycloneDX.Metadata.Component.PackageURL,
		BomRef:           report.CycloneDX.Metadata.Component.BOMRef,
		SerialNumber:     report.CycloneDX.SerialNumber,
		CycloneDxVersion: report.CycloneDX.Version,
		BomFormat:        report.CycloneDX.BOMFormat,
	}
	metricsJson, err := json.Marshal(metrics)
	if err!=nil {
		log.Println("error occurred while marshalling sbom metrics in agent", err.Error())
		return err
	}
	_, err = js.Publish(constants.TRIVY_SBOM_SUBJECT, metricsJson)
	if err != nil {
		return err
	}
	log.Printf("Trivy sbom report with Id %v has been published\n", metrics.ID)

}
}

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
			continue // Move on to the next image in case of an error
		}
		if out == nil {
			log.Printf("Trivy output is nil for image sbom %s", image.PullableImage)
			continue
		}
		// Check if the output is empty or invalid JSON
		if len(out) == 0 {
			log.Printf("Trivy output is empty for image sbom %s", image.PullableImage)
			continue // Move on to the next image
		}

		var report cyclonedx.BOM
		err = json.Unmarshal(out, &report)
		if err != nil {
			log.Printf("Error unmarshaling JSON data for image sbom %s: %v", image.PullableImage, err)
			continue // Move on to the next image in case of an error
		}
		publishTrivySbomReport(report, js)
	}
	return nil
}
