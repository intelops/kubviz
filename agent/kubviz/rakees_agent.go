package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/intelops/kubviz/agent/kubviz/rakkess"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const eventSubject_rakees = "METRICS.rakees"

func accessToOutcome(access rakkess.Access) (rakkess.Outcome, error) {
	switch access {
	case 0:
		return rakkess.None, nil
	case 1:
		return rakkess.Up, nil
	case 2:
		return rakkess.Down, nil
	case 3:
		return rakkess.Err, nil
	default:
		return rakkess.None, fmt.Errorf("unknown access code: %d", access)
	}
}

func RakeesOutput(config *rest.Config, js nats.JetStreamContext, wg *sync.WaitGroup, errCh chan error) {
	// Create a new Kubernetes client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		errCh <- err
	}

	// Retrieve all available resource types
	resourceList, err := client.Discovery().ServerPreferredResources()
	if err != nil {
		errCh <- err
	}
	var opts = rakkess.NewRakkessOptions()
	opts.Verbs = []string{"list", "create", "update", "delete"}
	opts.OutputFormat = "icon-table"
	opts.ResourceList = resourceList

	ctx, cancel := context.WithCancel(context.Background())
	catchCtrlC(cancel)

	res, err := rakkess.Resource(ctx, opts)
	if err != nil {
		fmt.Println("Error")
		errCh <- err
	}
	fmt.Println("Result..")
	for resourceType, access := range res {
		createOutcome, err := accessToOutcome(access["create"])
		if err != nil {
			errCh <- err
		}
		deleteOutcome, err := accessToOutcome(access["delete"])
		if err != nil {
			errCh <- err
		}
		listOutcome, err := accessToOutcome(access["list"])
		if err != nil {
			errCh <- err
		}
		updateOutcome, err := accessToOutcome(access["update"])
		if err != nil {
			errCh <- err
		}
		metrics := model.RakeesMetrics{
			ClusterName: ClusterName,
			Name:        resourceType,
			Create:      rakkess.HumanreadableAccessCode(createOutcome),
			Delete:      rakkess.HumanreadableAccessCode(deleteOutcome),
			List:        rakkess.HumanreadableAccessCode(listOutcome),
			Update:      rakkess.HumanreadableAccessCode(updateOutcome),
		}
		metricsJson, _ := json.Marshal(metrics)
		_, err = js.Publish(eventSubject_rakees, metricsJson)
		if err != nil {
			errCh <- err
		}
		log.Printf("Metrics with resource %s has been published", resourceType)
	}
	// t := res.Table(opts.Verbs)
	// t.Render(opts.Streams.Out, opts.OutputFormat)

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
