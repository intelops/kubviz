package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/kube-tarian/kubviz/constants"
	"github.com/kube-tarian/kubviz/model"
	"github.com/nats-io/nats.go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	exec "os/exec"
	"sync"
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
	for _, n := range nsList.Items {
		log.Printf("Publishing kube-score recommendations for namespace: %s\n", n.Name)
		publish(n.Name, js, errCh)
	}
}

func publish(ns string, js nats.JetStreamContext, errCh chan error) {
	cmd := "kubectl api-resources --verbs=list --namespaced -o name | xargs -n1 -I{} sh -c \"kubectl get {} -n " + ns + " -oyaml && echo ---\" | kube-score score - "
	log.Printf("Command:  %#v,", cmd)
	out, err := executeCommand(cmd)
	if err != nil {
		log.Println("Error occurred while running kube-score: ", err)
		errCh <- err
	}
	err = publishKubescoreMetrics(uuid.New().String(), ns, out, js)
	if err != nil {
		errCh <- err
	}
	errCh <- nil
}

func publishKubescoreMetrics(id string, ns string, recommendations string, js nats.JetStreamContext) error {
	metrics := model.KubeScoreRecommendations{
		ID:              id,
		Namespace:       ns,
		Recommendations: recommendations,
		ClusterName:     ClusterName,
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.SUBJECT, metricsJson)
	if err != nil {
		return err
	}
	log.Printf("Recommendations with ID:%s has been published\n", id)
	log.Printf("Recommendations  :%#v", recommendations)
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
