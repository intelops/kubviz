package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"log"
	exec "os/exec"
	"strings"

	"github.com/aquasecurity/trivy/pkg/k8s/report"
	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
)

func executeCommandTrivy(command string) ([]byte, error) {
	cmd := exec.Command("/bin/sh", "-c", command)
	var outc, errc bytes.Buffer
	cmd.Stdout = &outc
	cmd.Stderr = &errc

	err := cmd.Run()

	if err != nil {
		log.Println("Execute Trivy Command Error", err.Error())
	}

	return outc.Bytes(), err
}

// Compress data using gzip
func compressData(data []byte) ([]byte, error) {
	var compressedData bytes.Buffer
	gz := gzip.NewWriter(&compressedData)
	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return compressedData.Bytes(), nil
}

func RunTrivyK8sClusterScan(js nats.JetStreamContext) error {
	var report report.ConsolidatedReport
	cmdString := "trivy k8s --report summary cluster --exclude-nodes kubernetes.io/arch:amd64 --timeout 40m -f json --cache-dir /tmp/.cache"

	// Log the command before execution
	log.Printf("Executing command: %s\n", cmdString)

	// Execute the command
	out, err := executeCommandTrivy(cmdString)

	// Handle errors and process the command output as needed
	if err != nil {
		log.Printf("Error executing command: %v\n", err)
	}
	// Log the command output for debugging purposes
	log.Printf("Command output: %s\n", out)
	outStr := string(out)
	parts := strings.SplitN(outStr, "{", 2)
	if len(parts) <= 1 {
		log.Println("No output from k8s cluster scan command", err)
		return err
	}
	// log.Println("Command logs for k8s cluster scan", parts[0])
	jsonPart := "{" + parts[1]
	// log.Println("First 200 k8s cluster scan lines output", jsonPart[:200])
	// log.Println("Last 200 k8s cluster scan lines output", jsonPart[len(jsonPart)-200:])
	err = json.Unmarshal([]byte(jsonPart), &report)
	if err != nil {
		log.Printf("Error occurred while Unmarshalling json for k8s cluster scan: %v", err)
		return err
	}

	// Compress the Trivy scan report data
	compressedReport, err := compressData([]byte(jsonPart))
	if err != nil {
		log.Printf("Error compressing Trivy scan report: %v", err)
		return err
	}

	// Create a new TrivyReport struct with all the data
	trivyReport := model.Trivy{
		ID:                 uuid.New().String(),
		ClusterName:        ClusterName,
		Report:             report,
		CompressedReport:   compressedReport,
		UncompressedReport: []byte(jsonPart),
	}

	// Publish the TrivyReport
	err = publishTrivyK8sReport(trivyReport, js)
	if err != nil {
		return err
	}
	return nil
}

func publishTrivyK8sReport(trivyReport model.Trivy, js nats.JetStreamContext) error {
	// Create a JSON message for the TrivyReport
	metricsJson, _ := json.Marshal(trivyReport)

	// Publish the JSON message to the specified NATS subject
	_, err := js.Publish(constants.TRIVY_K8S_SUBJECT, metricsJson)
	if err != nil {
		return err
	}
	log.Printf("Trivy k8s cluster report with ID:%s has been published\n", trivyReport.ID)
	return nil
}
