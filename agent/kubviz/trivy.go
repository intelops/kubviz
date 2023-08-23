package main

import (
	"bytes"
	"encoding/json"
	"log"
	exec "os/exec"
	"sync"

	"github.com/aquasecurity/trivy/pkg/k8s/report"
	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
)

// func RunTrivyK8sClusterScan(wg *sync.WaitGroup, js nats.JetStreamContext, errCh chan error) {
// 	log.Println("*****started cluster scan")
// 	defer wg.Done()
// 	var report report.ConsolidatedReport
// 	out, err := executeCommand("trivy k8s --report summary cluster --timeout 60m -f json -q --cache-dir /tmp/.cache")
// 	parts := strings.SplitN(out, "{", 2)
// 	if len(parts) <= 1 {
// 		log.Println("No output from command", err)
// 		errCh <- err
// 		return
// 	}
// 	log.Println("Command logs", parts[0])
// 	jsonPart := "{" + parts[1]
// 	log.Println("First 200 lines output", jsonPart[:200])
// 	log.Println("Last 200 lines output", jsonPart[len(jsonPart)-200:])
// 	err = json.Unmarshal([]byte(jsonPart), &report)
// 	if err != nil {
// 		log.Printf("Error occurred while Unmarshalling json: %v", err)
// 		errCh <- err
// 	}
// 	publishTrivyK8sReport(report, js, errCh)
// }

func RunTrivyK8sClusterScan(wg *sync.WaitGroup, js nats.JetStreamContext, errCh chan error) {
	log.Println("*****started cluster scan")
	defer wg.Done()

	command := "trivy k8s --report summary cluster --timeout 60m -f json -q --cache-dir /tmp/.cache"
	out, err := executeCommandK8(command)

	if err != nil {
		log.Println("Error executing Trivy k8", err)
		// Move on to the next image in case of an error
	}

	var report report.ConsolidatedReport
	err = json.Unmarshal(out, &report)
	if err != nil {
		log.Printf("Error unmarshaling JSON data for k8 cluster %v", err)
		// Move on to the next image in case of an error
	}
	log.Println("report", report)
	publishTrivyK8sReport(report, js, errCh)
}

func publishTrivyK8sReport(report report.ConsolidatedReport, js nats.JetStreamContext, errCh chan error) {
	metrics := model.Trivy{
		ID:          uuid.New().String(),
		ClusterName: ClusterName,
		Report:      report,
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.TRIVY_K8S_SUBJECT, metricsJson)
	if err != nil {
		errCh <- err
	}
	log.Printf("Trivy report with ID:%s has been published\n", metrics.ID)
	errCh <- nil
}

func executeCommandK8(command string) ([]byte, error) {
	cmd := exec.Command("/bin/sh", "-c", command)
	//cmd := exec.Command(command)
	var outc, errc bytes.Buffer
	cmd.Stdout = &outc
	cmd.Stderr = &errc
	log.Println("*******before ece command")
	//stdout, err := cmd.Output()
	err := cmd.Run()
	log.Println("*******command ececuted")

	if err != nil {
		log.Println("Execute Command Error", err.Error())
	}
	log.Println("*******output", outc.String(), errc.String())

	return outc.Bytes(), err
}
