package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/intelops/kubviz/constants"

	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func PublishAllResources(result model.Resource, js nats.JetStreamContext) error {
	metrics := result
	metrics.ClusterName = ClusterName
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.EventSubject_getall_resource, metricsJson)
	if err != nil {
		return err
	}
	log.Printf("Metrics with resource %s in the %s namespace has been published", result.Resource, result.Namespace)
	return nil
}

func GetAllResources(config *rest.Config, js nats.JetStreamContext) error {
	// TODO: upto this uncomment for production
	// Create a new discovery client to discover all resources in the cluster
	dc := discovery.NewDiscoveryClientForConfigOrDie(config)

	// Create a new dynamic client to list resources in the cluster
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return err
	}
	// Get a list of all available API groups and versions in the cluster
	resourceLists, err := dc.ServerPreferredResources()
	if err != nil {
		return err
	}
	gvrs, err := discovery.GroupVersionResources(resourceLists)
	if err != nil {
		return err
	}
	// Iterate over all available API groups and versions and list all resources in each group
	for gvr := range gvrs {
		// List all resources in the group
		list, err := dynamicClient.Resource(gvr).Namespace("").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			// fmt.Printf("Error listing %s: %v\n", gvr.String(), err)
			continue
		}

		for _, item := range list.Items {
			age := time.Since(item.GetCreationTimestamp().Time).Round(time.Second).String()
			var resource model.Resource
			if item.GetNamespace() == "" {
				resource = model.Resource{
					Resource:  item.GetName(),
					Kind:      item.GetKind(),
					Namespace: "Default",
					Age:       age,
				}
			} else {
				resource = model.Resource{
					Resource:  item.GetName(),
					Kind:      item.GetKind(),
					Namespace: item.GetNamespace(),
					Age:       age,
				}

			}
			err := PublishAllResources(resource, js)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
