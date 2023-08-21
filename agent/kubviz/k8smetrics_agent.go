package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/nats-io/nats.go"

	"context"

	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"fmt"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	//  _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// constants for jetstream

type RuningEnv int

const (
	Development RuningEnv = iota
	Production
)

// env variables for getting
// nats token, natsurl, clustername
var (
	ClusterName string = os.Getenv("CLUSTER_NAME")
	token       string = os.Getenv("NATS_TOKEN")
	natsurl     string = os.Getenv("NATS_ADDRESS")
	//for local testing provide the location of kubeconfig
	// inside the civo file paste your kubeconfig
	// uncomment this line from Dockerfile.Kubviz (COPY --from=builder /workspace/civo /etc/myapp/civo)
	cluster_conf_loc      string = os.Getenv("CONFIG_LOCATION")
	schedulingIntervalStr string = os.Getenv("SCHEDULING_INTERVAL")
)

func main() {
	env := Production
	clusterMetricsChan := make(chan error, 1)
	var (
		wg        sync.WaitGroup
		config    *rest.Config
		clientset *kubernetes.Clientset
	)
	// connecting with nats ...
	nc, err := nats.Connect(natsurl, nats.Name("K8s Metrics"), nats.Token(token))
	checkErr(err)
	// creating a jetstream connection using the nats connection
	js, err := nc.JetStream()
	checkErr(err)
	// creating a stream with stream name METRICS
	err = createStream(js)
	checkErr(err)
	//setupAgent()
	if env != Production {
		config, err = clientcmd.BuildConfigFromFlags("", cluster_conf_loc)
		if err != nil {
			log.Fatal(err)
		}
		clientset = getK8sClient(config)
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
		clientset = getK8sClient(config)
	}

	// starting the endless go routine to monitor the cluster
	go publishMetrics(clientset, js, clusterMetricsChan)

	// starting all the go routines
	collectAndPublishMetrics := func() {
		// error channels declared for the go routines
		outdatedErrChan := make(chan error, 1)
		kubePreUpgradeChan := make(chan error, 1)
		getAllResourceChan := make(chan error, 1)
		trivyK8sMetricsChan := make(chan error, 1)
		kubescoreMetricsChan := make(chan error, 1)
		trivyImagescanChan := make(chan error, 1)
		trivySbomcanChan := make(chan error, 1)
		RakeesErrChan := make(chan error, 1)
		// Start a goroutine to handle errors
		doneChan := make(chan bool)
		go func() {
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
				case err := <-trivyImagescanChan:
					if err != nil {
						log.Println(err)
					}
				case err := <-trivySbomcanChan:
					if err != nil {
						log.Println(err)
					}
				case err := <-trivyK8sMetricsChan:
					if err != nil {
						log.Println(err)
					}
				case err := <-RakeesErrChan:
					if err != nil {
						log.Println(err)
					}
				case <-doneChan:
					return // All other goroutines have finished, so exit the goroutine
				}
			}
		}()
		wg.Add(1) // Initialize the WaitGroup for the seven goroutines
		// ... start other goroutines ...
		//go outDatedImages(config, js, &wg, outdatedErrChan)
		//go KubePreUpgradeDetector(config, js, &wg, kubePreUpgradeChan)
		//go GetAllResources(config, js, &wg, getAllResourceChan)
		//go RakeesOutput(config, js, &wg, RakeesErrChan)
		//go getK8sEvents(clientset)
		//go RunTrivyImageScans(config, js, &wg, trivyImagescanChan)
		//go RunKubeScore(clientset, js, &wg, kubescoreMetricsChan)
		//go RunTrivyK8sClusterScan(&wg, js, trivyK8sMetricsChan)
		go RunTrivyScans(config, js, &wg, trivySbomcanChan,trivyImagescanChan,trivyK8sMetricsChan)

		wg.Wait()
		// once the go routines completes we will close the error channels
		//close(outdatedErrChan)
		//close(kubePreUpgradeChan)
		//close(getAllResourceChan)
		// close(clusterMetricsChan)
		//close(kubescoreMetricsChan)
		close(trivyImagescanChan)
		close(trivySbomcanChan)
		close(trivyK8sMetricsChan)
		//close(RakeesErrChan)
		// Signal that all other goroutines have finished
		doneChan <- true
		close(doneChan)
	}
	collectAndPublishMetrics()
	if schedulingIntervalStr == "" {
		schedulingIntervalStr = "20m" // Default value, e.g., 20 minutes
	}
	schedulingInterval, err := time.ParseDuration(schedulingIntervalStr)
	if err != nil {
		log.Fatalf("Failed to parse SCHEDULING_INTERVAL: %v", err)
	}
	s := gocron.NewScheduler(time.UTC)
	s.Every(schedulingInterval).Do(collectAndPublishMetrics) // Run immediately and then at the scheduled interval
	s.StartBlocking()                                        // Blocks the main function
}

func RunTrivyScans(config *rest.Config, js nats.JetStreamContext, wg *sync.WaitGroup,trivySbomcanChan chan error, trivyImagescanChan chan error, trivyK8sMetricsChan chan error) {
	defer wg.Done()
	RunTrivySbomScan(config, js, wg, trivySbomcanChan)
	RunTrivyImageScans(config, js, wg, trivyImagescanChan)
	RunTrivyK8sClusterScan(wg, js, trivyK8sMetricsChan)
	
}

// publishMetrics publishes stream of events
// with subject "METRICS.created"
func publishMetrics(clientset *kubernetes.Clientset, js nats.JetStreamContext, errCh chan error) {
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
	_, err := js.Publish(constants.EventSubject, metricsJson)
	if err != nil {
		return true, err
	}
	log.Printf("Metrics with ID:%s has been published\n", id)
	return false, nil
}

// createStream creates a stream by using JetStreamContext
func createStream(js nats.JetStreamContext) error {
	// Check if the METRICS stream already exists; if not, create it.
	stream, err := js.StreamInfo(constants.StreamName)
	log.Printf("Retrieved stream %s", fmt.Sprintf("%v", stream))
	if err != nil {
		log.Printf("Error getting stream %s", err)
	}
	if stream == nil {
		log.Printf("creating stream %q and subjects %q", constants.StreamName, constants.StreamSubjects)
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     constants.StreamName,
			Subjects: []string{constants.StreamSubjects},
		})
		checkErr(err)
	}
	return nil

}

func getK8sClient(config *rest.Config) *kubernetes.Clientset {
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
