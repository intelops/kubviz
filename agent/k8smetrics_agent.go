package main

import (
	"encoding/json"
	"github.com/kube-tarian/kubviz/agent/config"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kube-tarian/kubviz/model"
	"github.com/nats-io/nats.go"

	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	"fmt"

	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	_ "k8s.io/client-go/plugin/pkg/client/auth/openstack"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

// constants for jetstream
const (
	streamName     = "METRICS"
	streamSubjects = "METRICS.*"
	eventSubject   = "METRICS.event"
	allSubject     = "METRICS.all"
)

// env variables for getting
// nats token, natsurl, clustername
var (
	ClusterName string = os.Getenv("CLUSTER_NAME")
	token       string = os.Getenv("NATS_TOKEN")
	natsurl     string = os.Getenv("NATS_ADDRESS")
)

func main() {
	// error channels declared for the go routines
	outdatedErrChan := make(chan error, 1)
	kubePreUpgradeChan := make(chan error, 1)
	getAllResourceChan := make(chan error, 1)
	clusterMetricsChan := make(chan error, 1)
	kubescoreMetricsChan := make(chan error, 1)
	var wg sync.WaitGroup
	// waiting for 4 go routines
	wg.Add(5)
	// connecting with nats ...
	nc, err := nats.Connect(natsurl, nats.Name("K8s Metrics"), nats.Token(token))
	checkErr(err)
	// creating a jetstream connection using the nats connection
	js, err := nc.JetStream()
	checkErr(err)
	// creating a stream with stream name METRICS
	err = createStream(js)
	checkErr(err)
	// getting kubernetes clientset
	clientset := getK8sClient()

	setupAgent()

	// starting all the go routines
	go outDatedImages(js, &wg, outdatedErrChan)
	go KubePreUpgradeDetector(js, &wg, kubePreUpgradeChan)
	go GetAllResources(js, &wg, getAllResourceChan)
	getK8sEvents(clientset)
	go publishMetrics(clientset, js, &wg, clusterMetricsChan)
	go RunKubeScore(clientset, js, &wg, kubescoreMetricsChan)
	wg.Wait()
	// once the go routines completes we will close the error channels
	close(outdatedErrChan)
	close(kubePreUpgradeChan)
	close(getAllResourceChan)
	close(clusterMetricsChan)
	close(kubescoreMetricsChan)
	// for loop will wait for the error channels
	// logs if any error occurs
	for {
		select {
		case err := <-outdatedErrChan:
			if err != nil {
				log.Println(err)
			}
		case err := <-kubePreUpgradeChan:
			if err != nil {
				log.Println(err)
			}
		case err := <-getAllResourceChan:
			if err != nil {
				log.Println(err)
			}
		case err := <-clusterMetricsChan:
			if err != nil {
				log.Println(err)
			}
		case err := <-kubescoreMetricsChan:
			if err != nil {
				log.Println(err)
			}
		}
	}

}

func setupAgent() {
	configurations, err := config.GetAgentConfigurations()
	if err != nil {
		log.Printf("Failed to get agent config: %v\n", err)
		panic(err)
	}
	k8s := &K8sData{
		Namespace:          configurations.SANamespace,
		ServiceAccountName: configurations.SAName,
		KubeconfigFileName: KUBECONFIG,
	}
	_, err = k8s.GenerateKubeConfiguration()
	if err != nil {
		log.Printf("Failed to generate kubeconfig: %v\n", err)
		panic(err)
	}
}

// publishMetrics publishes stream of events
// with subject "METRICS.created"
func publishMetrics(clientset *kubernetes.Clientset, js nats.JetStreamContext, wg *sync.WaitGroup, errCh chan error) {
	defer wg.Done()
	//Publish Nodes data
	// for i := 1; i <= 10; i++ {
	// 	shouldReturn, returnValue := publishK8sMetrics(i, "Node", getK8sNodes(clientset), js)
	// 	if shouldReturn {
	// 		return returnValue
	// 	}
	// 	time.Sleep(100 * time.Millisecond)
	// }
	//Publish Pods data
	// for i := 1; i <= 10; i++ {

	// 	shouldReturn, returnValue := publishK8sMetrics(i, "Pod", getK8sPods(clientset), js)
	// 	if shouldReturn {
	// 		return returnValue
	// 	}
	// 	time.Sleep(100 * time.Millisecond)
	// }
	//Publish events data
	//publishK8sMetrics(1, "Event", getK8sEvents(clientset), js)
	watchK8sEvents(clientset, js)

	errCh <- nil
}

func publishK8sMetrics(id string, mtype string, mdata *v1.Event, js nats.JetStreamContext) (bool, error) {
	metrics := model.Metrics{
		ID:          id,
		Type:        mtype,
		Event:       mdata,
		ClusterName: ClusterName,
	}
	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(eventSubject, metricsJson)
	if err != nil {
		return true, err
	}
	log.Printf("Metrics with ID:%s has been published\n", id)
	return false, nil
}

// createStream creates a stream by using JetStreamContext
func createStream(js nats.JetStreamContext) error {
	// Check if the METRICS stream already exists; if not, create it.
	stream, err := js.StreamInfo(streamName)
	log.Printf("Retrieved stream %s", fmt.Sprintf("%v", stream))
	if err != nil {
		log.Printf("Error getting stream %s", err)
	}
	if stream == nil {
		log.Printf("creating stream %q and subjects %q", streamName, streamSubjects)
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{streamSubjects},
		})
		checkErr(err)
	}
	return nil

}

func getK8sClient() *kubernetes.Clientset {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	checkErr(err)
	return clientset
}

func getK8sPods(clientset *kubernetes.Clientset) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	checkErr(err)
	var sb strings.Builder
	for i, pod := range pods.Items {
		sb.WriteString("Name-" + strconv.Itoa(i) + ": ")
		sb.WriteString(pod.Name)
		sb.WriteString("   ")
		sb.WriteString("Namespace-" + strconv.Itoa(i) + ": ")
		sb.WriteString(pod.Namespace)
		sb.WriteString("   ")
	}
	return sb.String()
}

func getK8sNodes(clientset *kubernetes.Clientset) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	checkErr(err)
	var sb strings.Builder
	for i, node := range nodes.Items {
		sb.WriteString("Name-" + strconv.Itoa(i) + ": ")
		sb.WriteString(node.Name)
	}
	return sb.String()
}

func getK8sEvents(clientset *kubernetes.Clientset) string {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	events, err := clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	checkErr(err)
	j, err := json.MarshalIndent(events, "", "  ")
	checkErr(err)
	log.Printf(string(j))
	return string(j)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func watchK8sEvents(clientset *kubernetes.Clientset, js nats.JetStreamContext) {
	watchlist := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"events",
		v1.NamespaceAll,
		fields.Everything(),
	)
	_, controller := cache.NewInformer( // also take a look at NewSharedIndexInformer
		watchlist,
		&v1.Event{},
		0, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				event := obj.(*v1.Event)
				fmt.Printf("Event namespace: %s \n", event.GetNamespace())
				// j, err := json.MarshalIndent(obj, "", "  ")
				// checkErr(err)
				//fmt.Printf("Add event: %s \n", event)
				y, err := yaml.Marshal(event)
				if err != nil {
					fmt.Printf("err: %v\n", err)
				}
				fmt.Printf("Add event: %s \n", y)
				publishK8sMetrics(string(event.ObjectMeta.UID), "ADD", event, js)
			},
			DeleteFunc: func(obj interface{}) {
				event := obj.(*v1.Event)
				fmt.Printf("Delete event: %s \n", obj)
				publishK8sMetrics(string(event.ObjectMeta.UID), "DELETE", event, js)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				event := newObj.(*v1.Event)
				fmt.Printf("Change event \n")
				publishK8sMetrics(string(event.ObjectMeta.UID), "UPDATE", event, js)
			},
		},
	)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}
