package main

import (
	"context"
	"encoding/json"
	"log"
	exec "os/exec"

	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func RunKubeScore(clientset *kubernetes.Clientset, js nats.JetStreamContext) error {

	ctx:=context.Background()
	tracer := otel.Tracer("kubescore")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "RunKubeScore")
	span.SetAttributes(attribute.String("kubescore-run", "kubescore-output"))
	defer span.End()
	
	nsList, err := clientset.CoreV1().
		Namespaces().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println("Error occurred while getting client set for kube-score: ", err)
		return err
	}

	log.Printf("Namespace size: %d", len(nsList.Items))
	for _, n := range nsList.Items {
		log.Printf("Publishing kube-score recommendations for namespace: %s\n", n.Name)
		publish(n.Name, js)
	}
	return nil
}

func publish(ns string, js nats.JetStreamContext) error {
	cmd := "kubectl api-resources --verbs=list --namespaced -o name | xargs -n1 -I{} sh -c \"kubectl get {} -n " + ns + " -oyaml && echo ---\" | kube-score score - "
	log.Printf("Command:  %#v,", cmd)
	out, err := executeCommand(cmd)
	if err != nil {
		log.Println("Error occurred while running kube-score: ", err)
		return err
	}
	err = publishKubescoreMetrics(uuid.New().String(), ns, out, js)
	if err != nil {
		return err
	}
	return nil
}

func publishKubescoreMetrics(id string, ns string, recommendations string, js nats.JetStreamContext) error {
	metrics := model.KubeScoreRecommendations{
		ID:              id,
		Namespace:       ns,
		Recommendations: recommendations,
		ClusterName:     ClusterName,
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.KUBESCORE_SUBJECT, metricsJson)
	if err != nil {
		return err
	}
	log.Printf("Recommendations with ID:%s has been published\n", id)
	log.Printf("Recommendations  :%#v", recommendations)
	return nil
}

func executeCommand(command string) (string, error) {

	ctx:=context.Background()
	tracer := otel.Tracer("kubescore")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "executeCommand")
	span.SetAttributes(attribute.String("kubescore", "kubescore-command-running"))
	defer span.End()

	cmd := exec.Command("/bin/sh", "-c", command)
	stdout, err := cmd.Output()

	if err != nil {
		log.Println("Execute Command Error", err.Error())
	}

	// Print the output
	log.Println(string(stdout))
	return string(stdout), nil
}
