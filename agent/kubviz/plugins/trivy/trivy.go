package trivy

import (
	"bytes"
	//	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	exec "os/exec"
	"strings"

	"github.com/aquasecurity/trivy/pkg/k8s/report"
	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"

	//	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/nats-io/nats.go"
	// "go.opentelemetry.io/otel"
	// "go.opentelemetry.io/otel/attribute"
)

var ClusterName string = os.Getenv("CLUSTER_NAME")

func executeCommandTrivy(command string) ([]byte, error) {

	// ctx := context.Background()
	// tracer := otel.Tracer("trivy-cluster")
	// _, span := tracer.Start(opentelemetry.BuildContext(ctx), "executeCommandTrivy")
	// span.SetAttributes(attribute.String("trivy-k8s-agent", "command-running"))
	// defer span.End()

	cmd := exec.Command("/bin/sh", "-c", command)
	var outc, errc bytes.Buffer
	cmd.Stdout = &outc
	cmd.Stderr = &errc
	log.Println("&&&&reason for kill&&&", errc)
	err := cmd.Run()

	if err != nil {
		log.Println("Execute Trivy Command Error", err.Error())
	}

	return outc.Bytes(), err
}
func RunTrivyK8sClusterScan(js nats.JetStreamContext) error {
	pvcMountPath := "/mnt/agent/kbz"
	trivyCacheDir := fmt.Sprintf("%s/trivy-cache", pvcMountPath)
	err := os.MkdirAll(trivyCacheDir, 0755)
	if err != nil {
		log.Printf("Error creating Trivy cache directory: %v\n", err)
		return err
	}
	var report report.ConsolidatedReport

	// ctx := context.Background()
	// tracer := otel.Tracer("trivy-cluster")
	// _, span := tracer.Start(opentelemetry.BuildContext(ctx), "RunTrivyK8sClusterScan")
	// span.SetAttributes(attribute.String("cluster-name", report.ClusterName))
	// defer span.End()

	cmdString := fmt.Sprintf("trivy k8s --report=summary cluster --exclude-nodes kubernetes.io/arch:amd64 --timeout 60m -f json --cache-dir %s --debug", trivyCacheDir)
	// clearCacheCmd := "trivy k8s --clear-cache"
	out, err := executeCommandTrivy(cmdString)
	if err != nil {
		log.Printf("Error executing command: %v\n", err)
		return err
	}
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
	// _, err = executeCommandTrivy(clearCacheCmd)
	// if err != nil {
	// 	log.Printf("Error executing command: %v\n", err)
	// 	return err
	// }
	err = PublishTrivyK8sReport(report, js)
	if err != nil {
		return err
	}
	return nil
}

func PublishTrivyK8sReport(report report.ConsolidatedReport, js nats.JetStreamContext) error {
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
	log.Printf("Trivy k8s cluster report with ID:%s has been published\n", metrics.Report)
	return nil
}
