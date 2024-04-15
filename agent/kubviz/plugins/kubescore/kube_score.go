package kubescore

import (
	"context"
	"encoding/json"
	"log"
	"os"
	exec "os/exec"

	"github.com/google/uuid"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/nats-io/nats.go"
	"github.com/zegl/kube-score/renderer/json_v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var ClusterName string = os.Getenv("CLUSTER_NAME")

func RunKubeScore(clientset *kubernetes.Clientset, js nats.JetStreamContext) error {
	nsList, err := clientset.CoreV1().
		Namespaces().
		List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Println("Error occured while getting client set for kube-score: ", err)
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
	var report []json_v2.ScoredObject
	cmd := "kubectl api-resources --verbs=list --namespaced -o name | xargs -n1 -I{} sh -c \"kubectl get {} -n " + ns + " -oyaml && echo ---\" | kube-score score - -o json"
	log.Printf("Command:  %#v,", cmd)
	out, err := ExecuteCommand(cmd)
	if err != nil {
		log.Println("Error occured while running kube-score: ", err)
		return err
	}
	// 	// Continue with the rest of the code...
	err = json.Unmarshal([]byte(out), &report)
	if err != nil {
		log.Printf("Error occured while Unmarshalling json: %v", err)
		return err
	}

	publishKubescoreMetrics(report, js)
	//err = publishKubescoreMetrics(uuid.New().String(), ns, out, js)
	return nil
}

func publishKubescoreMetrics(report []json_v2.ScoredObject, js nats.JetStreamContext) error {

	ctx := context.Background()
	tracer := otel.Tracer("kubescore")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "publishKubescoreMetrics")
	span.SetAttributes(attribute.String("kubescore-plugin-agent", "kubescore-output"))
	defer span.End()

	metrics := model.KubeScoreRecommendations{
		ID:          uuid.New().String(),
		ClusterName: ClusterName,
		Report:      report,
	}
	metricsJSON, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.KUBESCORE_SUBJECT, metricsJSON)
	if err != nil {
		return err
	}
	//log.Printf("Recommendations with ID:%s has been published\n", id)
	log.Printf("Recommendations  :%#v", report)
	return nil
}

func ExecuteCommand(command string) (string, error) {

	ctx := context.Background()
	tracer := otel.Tracer("kubescore")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "executeCommand")
	span.SetAttributes(attribute.String("kubescore-agent", "kubescore-command-running"))
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
