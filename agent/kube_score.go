package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	//checkErr(err)
	fmt.Println(err)

	for _, n := range nsList.Items {
		log.Println("Publishing kube-score recommendations for namespace: ", n.Namespace)
		publish(n.Namespace, js, errCh)
	}
}

func publish(ns string, js nats.JetStreamContext, errCh chan error) {
	out, err := executeCommand("kubectl api-resources --verbs=list --namespaced -o name | xargs -n1 -I{} bash -c \"kubectl get {} -n " + ns + " -oyaml && echo ---\" | kube-score score - " +
		" --kubeconfig=" + constants.KUBECONFIG)
	if err != nil {
		log.Println("Error occurred while running kube-score: ", err)
		errCh <- err
	}
	err = publishKubescoreMetrics(uuid.New().String(), "all", out, js)
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
	return nil
}

func executeCommand(command string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", command)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	// Print the output
	log.Println(string(stdout))
	return string(stdout), nil
}