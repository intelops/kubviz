package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aquasecurity/trivy/pkg/k8s/report"
	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func RunTrivyK8sClusterScan(clientset *kubernetes.Clientset, js nats.JetStreamContext) error {

	namespaceList, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println("Error occurred while listing namespaces: ", err)
		return err
	}

	for _, ns := range namespaceList.Items {
		namespace := ns.Name
		log.Printf("Scanning namespace: %s\n", namespace)

		var report report.ConsolidatedReport
		cmd := fmt.Sprintf("trivy k8s --namespace %s --report summary all --timeout 60m -f json -q --cache-dir /tmp/.cache", namespace)
		out, err := executeCommand(cmd)
		if err != nil {
			log.Printf("Error occurred while running Trivy scan for namespace %s: %v", namespace, err)
			continue // Continue to the next namespace on error.
		}

		parts := strings.SplitN(out, "{", 2)
		if len(parts) <= 1 {
			log.Printf("No output from Trivy scan command for namespace %s\n", namespace)
			continue // Continue to the next namespace if there's no output.
		}

		jsonPart := "{" + parts[1]
		err = json.Unmarshal([]byte(jsonPart), &report)
		if err != nil {
			log.Printf("Error occurred while Unmarshalling JSON for namespace %s: %v", namespace, err)
			continue // Continue to the next namespace on error.
		}

		err = publishTrivyK8sReport(report, js)
		if err != nil {
			log.Printf("Error occurred while publishing Trivy scan report for namespace %s: %v", namespace, err)
		}
		return nil
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
