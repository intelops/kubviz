package main

import (
	"encoding/json"
	"fmt"
	"log"
	exec "os/exec"

	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	"github.com/zegl/kube-score/renderer/json_v2"
	"k8s.io/client-go/rest"
)

func RunKubeScore(config *rest.Config, js nats.JetStreamContext) error {
	// _, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	log.Printf("Error creating Kubernetes clientset: %v", err)
	// 	return err
	// }
	//defer wg.Done()
	var report []json_v2.ScoredObject
	cmd := fmt.Sprintf(`kubectl api-resources --verbs=list --namespaced -o name | xargs -n1 -I{} sh -c "kubectl get {} --all-namespaces -oyaml && echo ---" | kube-score score - -o json`)
	log.Printf("Command: %s", cmd)

	// Execute the command
	out, err := executeCommand(cmd)
	if err != nil {
		log.Printf("Error executing command: %s", err)
		return err
	}

	// Log the output of the kubectl command
	log.Printf("kubectl Command Output: %s", out)

	// Continue with the rest of the code...
	err = json.Unmarshal([]byte(out), &report)
	if err != nil {
		log.Printf("Error occurred while Unmarshalling json: %v", err)
		return err
	}

	publishKubescoreMetrics(report, js)
	return nil
}

func publishKubescoreMetrics(report []json_v2.ScoredObject, js nats.JetStreamContext) {
	metrics := model.KubeScoreRecommendations{
		ID:          uuid.New().String(),
		ClusterName: ClusterName,
		Report:      report,
	}
	metricsJson, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error marshaling metrics to JSON: %v", err)
		return
	}
	_, err = js.Publish(constants.KUBESCORE_SUBJECT, metricsJson)
	if err != nil {
		log.Printf("error occures while publish %v", err)
		return
	}
}

func executeCommand(command string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", command)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println("Execute Command Error", err.Error())
	}

	// Print the output
	log.Println(string(stdout))
	return string(stdout), nil
}
