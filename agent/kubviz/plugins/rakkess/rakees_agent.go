package rakkess

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var ClusterName string = os.Getenv("CLUSTER_NAME")

func accessToOutcome(access Access) (Outcome, error) {
	switch access {
	case 0:
		return None, nil
	case 1:
		return Up, nil
	case 2:
		return Down, nil
	case 3:
		return Err, nil
	default:
		return None, fmt.Errorf("unknown access code: %d", access)
	}
}

func RakeesOutput(config *rest.Config, js nats.JetStreamContext) error {

	ctx := context.Background()
	tracer := otel.Tracer("rakees")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "RakeesOutput")
	span.SetAttributes(attribute.String("rakees-plugin-agent", "rakees-output"))
	defer span.End()

	// Create a new Kubernetes client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	// Retrieve all available resource types
	resourceList, err := client.Discovery().ServerPreferredResources()
	if err != nil {
		return err
	}
	var opts = NewRakkessOptions()
	opts.Verbs = []string{"list", "create", "update", "delete"}
	opts.OutputFormat = "icon-table"
	opts.ResourceList = resourceList

	ctx, cancel := context.WithCancel(context.Background())
	catchCtrlC(cancel)

	res, err := Resource(ctx, opts)
	if err != nil {
		fmt.Println("Error")
		return err
	}
	fmt.Println("Result..")
	for resourceType, access := range res {
		createOutcome, err := accessToOutcome(access["create"])
		if err != nil {
			return err
		}
		deleteOutcome, err := accessToOutcome(access["delete"])
		if err != nil {
			return err
		}
		listOutcome, err := accessToOutcome(access["list"])
		if err != nil {
			return err
		}
		updateOutcome, err := accessToOutcome(access["update"])
		if err != nil {
			return err
		}
		metrics := model.RakeesMetrics{
			ClusterName: ClusterName,
			Name:        resourceType,
			Create:      HumanreadableAccessCode(createOutcome),
			Delete:      HumanreadableAccessCode(deleteOutcome),
			List:        HumanreadableAccessCode(listOutcome),
			Update:      HumanreadableAccessCode(updateOutcome),
		}
		metricsJson, _ := json.Marshal(metrics)
		_, err = js.Publish(constants.EventSubject_rakees, metricsJson)
		if err != nil {
			return err
		}
		//log.Printf("Metrics with resource %s has been published", resourceType)
	}
	return nil

}

func catchCtrlC(cancel context.CancelFunc) {
	catchSigs(cancel, syscall.SIGINT, syscall.SIGPIPE, syscall.SIGTERM)
}

func catchSigs(cancel context.CancelFunc, sigs ...os.Signal) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)

	go func() {
		<-sigChan
		cancel()
	}()
}
