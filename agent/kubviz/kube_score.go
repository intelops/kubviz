package main

import (
	"context"
	"encoding/json"
	"log"
	exec "os/exec"
	"sync"

	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	"github.com/zegl/kube-score/renderer/json_v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func RunKubeScore(clientset *kubernetes.Clientset, js nats.JetStreamContext, wg *sync.WaitGroup, errCh chan error) {
	defer wg.Done()

	nsList, err := clientset.CoreV1().
		Namespaces().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println("Error occurred while getting client set for kube-score: ", err)
		return
	}

	log.Printf("Namespace size: %d", len(nsList.Items))
	var wgNamespaces sync.WaitGroup
	for _, n := range nsList.Items {
		wgNamespaces.Add(1)
		log.Printf("Publishing kube-score recommendations for namespace: %s\n", n.Name)
		go publish(n.Name, js, &wgNamespaces, errCh)
	}
}

func publish(ns string, js nats.JetStreamContext, wg *sync.WaitGroup, errCh chan error) {
	defer wg.Done()
	var report []json_v2.ScoredObject

	cmd := "kubectl api-resources --verbs=list --namespaced -o name | xargs -n1 -I{} sh -c \"kubectl get {} -n " + ns + " -oyaml && echo ---\" | kube-score score - -o json"
	log.Printf("Command:  %#v,", cmd)
	out, err := executeCommand(cmd)

	err = json.Unmarshal([]byte(out), &report)
	if err != nil {
		log.Printf("Error occurred while Unmarshalling json: %v", err)
		errCh <- err
	}

	if err != nil {
		log.Println("Error occurred while running kube-score: ", err)
		errCh <- err
	}
	err = publishKubescoreMetrics(uuid.New().String(), report, js)
	if err != nil {
		errCh <- err
	}
	errCh <- nil
}

func publishKubescoreMetrics(id string, report []json_v2.ScoredObject, js nats.JetStreamContext) error {
	metrics := model.KubeScoreRecommendations{
		ID:          id,
		ClusterName: ClusterName,
		Report:      report,
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.KUBESCORE_SUBJECT, metricsJson)
	if err != nil {
		return err
	}
	log.Printf("Recommendations with ID:%s has been published\n", id)
	log.Printf("Recommendations  :%#v", report)
	return nil
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
