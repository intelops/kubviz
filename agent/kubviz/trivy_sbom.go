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

	log.Printf("Trivy report with Id %v has been published\n", metrics.ID)
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
	pvcMountPath := "/mnt/agent/kbz"
	trivySbomCacheDir := fmt.Sprintf("%s/trivy-sbomcache", pvcMountPath)
	err := os.MkdirAll(trivySbomCacheDir, 0755)
	if err != nil {
		log.Printf("Error creating Trivy cache directory: %v\n", err)
		return err
	}
	// clearCacheCmd := "trivy image --clear-cache"

	log.Println("trivy sbom run started")
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
		// log.Println("report", report)
		// _, err = executeCommandTrivy(clearCacheCmd)
		// if err != nil {
		// 	log.Printf("Error executing command: %v\n", err)
		// 	return err
		// }
		// Publish the report using the given function
		publishTrivySbomReport(report, js)
	}
	return nil
}
