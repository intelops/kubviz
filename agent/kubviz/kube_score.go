package main

import (
	"encoding/json"
	"log"
	exec "os/exec"
	"sync"

	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	"github.com/zegl/kube-score/renderer/json_v2"
	"k8s.io/client-go/kubernetes"
)

func RunKubeScore(clientset *kubernetes.Clientset, js nats.JetStreamContext, wg *sync.WaitGroup, errCh chan error) {
	defer wg.Done()
	var report []json_v2.ScoredObject
	cmd := "kubectl api-resources --verbs=list --namespaced -o name | xargs -n1 -I{} sh -c \"kubectl get {} --all-namespaces -oyaml && echo ---\" | kube-score score - -o json"
	log.Printf("Command:  %#v,", cmd)
	out, err := executeCommand(cmd)

	if err != nil {
		log.Printf("Error scanning image %s:", err)
		//continue // Move on to the next image in case of an error
	}

	err = json.Unmarshal([]byte(out), &report)

	if err != nil {
		log.Printf("Error occurred while Unmarshalling json: %v", err)
		//continue // Move on to the next image in case of an error
	}
	publishKubescoreMetrics(report, js, errCh)
	//log.Println("Publishing kube-score recommendations for namespace")

	// If you want to publish the report or perform any other action with it, you can do it here

}

func publishKubescoreMetrics(report []json_v2.ScoredObject, js nats.JetStreamContext, errCh chan error) {
	metrics := model.KubeScoreRecommendations{
		ID:          uuid.New().String(),
		ClusterName: ClusterName,
		Report:      report,
	}
	metricsJson, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error marshaling metrics to JSON: %v", err)
		errCh <- err
		return
	}

	_, err = js.Publish(constants.KUBESCORE_SUBJECT, metricsJson)
	if err != nil {
		log.Printf("error occures while publish %v", err)
		errCh <- err
		return
	}
	log.Printf("KubeScore report with ID:%s has been published\n", metrics.ID)
	errCh <- nil
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
