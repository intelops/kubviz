package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/aquasecurity/trivy/pkg/k8s/report"
	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
)

func RunTrivyK8sClusterScan(js nats.JetStreamContext) error {
	var report report.ConsolidatedReport
	out, err := executeCommand("trivy k8s --report summary cluster --timeout 60m -f json -q --cache-dir /tmp/.cache")
	log.Println("Commnd for k8s cluster scan: trivy k8s --report summary cluster --timeout 60m -f json -q --cache-dir /tmp/.cache")
	parts := strings.SplitN(out, "{", 2)
	if len(parts) <= 1 {
		log.Println("No output from k8s cluster scan command", err)
		return err
	}
	log.Println("Command logs for k8s cluster scan", parts[0])
	jsonPart := "{" + parts[1]
	// log.Println("First 200 k8s cluster scan lines output", jsonPart[:200])
	// log.Println("Last 200 k8s cluster scan lines output", jsonPart[len(jsonPart)-200:])
	err = json.Unmarshal([]byte(jsonPart), &report)
	if err != nil {
		log.Printf("Error occurred while Unmarshalling json for k8s cluster scan: %v", err)
		return err
	}
	err = publishTrivyK8sReport(report, js)
	if err != nil {
		return err
	}
	return nil
}

func publishTrivyK8sReport(report report.ConsolidatedReport, js nats.JetStreamContext) error {
	metrics := model.Trivy{
		ID:          uuid.New().String(),
		ClusterName: ClusterName,
		Report:      report,
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.TRIVY_K8S_SUBJECT, metricsJson)
	if err != nil {
		return err
	}
	log.Printf("Trivy k8s cluster report with ID:%s has been published\n", metrics.ID)
	return nil
}
