package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/intelops/go-common/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/go-co-op/gocron"
	"github.com/nats-io/nats.go"

	"context"

	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/opentelemetry"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"fmt"

	"github.com/intelops/kubviz/agent/config"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	//  _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	"github.com/intelops/kubviz/agent/server"
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

	natsurl string = os.Getenv("NATS_ADDRESS")

	//for local testing provide the location of kubeconfig
	// inside the civo file paste your kubeconfig
	// uncomment this line from Dockerfile.Kubviz (COPY --from=builder /workspace/civo /etc/myapp/civo)
	cluster_conf_loc      string = os.Getenv("CONFIG_LOCATION")
	schedulingIntervalStr string = os.Getenv("SCHEDULING_INTERVAL")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	env := Production
	clusterMetricsChan := make(chan error, 1)
	cfg, err := config.GetAgentConfigurations()
	if err != nil {
		log.Fatal("Failed to retrieve agent configurations", err)
	}
	var (
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

	tp, err := opentelemetry.InitTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	go publishMetrics(clientset, js, clusterMetricsChan)
	go server.StartServer()
	collectAndPublishMetrics := func() {
		err := outDatedImages(config, js)
		LogErr(err)
		err = KubePreUpgradeDetector(config, js)
		LogErr(err)
		err = GetAllResources(config, js)
		LogErr(err)
		err = RakeesOutput(config, js)
		LogErr(err)
		// //getK8sEvents(clientset)
		err = RunTrivySbomScan(config, js)
		LogErr(err)
		err = RunTrivyImageScans(config, js)
		LogErr(err)
		err = RunTrivyK8sClusterScan(js)
		LogErr(err)
		err = RunKubeScore(clientset, js)
		LogErr(err)
	}

	collectAndPublishMetrics()

	if cfg.SchedulerEnable { // Assuming "cfg.Schedule" is a boolean indicating whether to schedule or not.
		scheduler := initScheduler(config, js, *cfg, clientset)

		// Start the scheduler
		scheduler.Start()
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		<-signals

		scheduler.Stop()
	} else {
		if schedulingIntervalStr == "" {
			schedulingIntervalStr = "20m"
		}
		schedulingInterval, err := time.ParseDuration(schedulingIntervalStr)
		if err != nil {
			log.Fatalf("Failed to parse SCHEDULING_INTERVAL: %v", err)
		}
		s := gocron.NewScheduler(time.UTC)
		s.Every(schedulingInterval).Do(func() {
			collectAndPublishMetrics()
		})
		s.StartBlocking()
	}
}

// publishMetrics publishes stream of events
// with subject "METRICS.created"
func publishMetrics(clientset *kubernetes.Clientset, js nats.JetStreamContext, errCh chan error) {

	ctx := context.Background()
	tracer := otel.Tracer("kubviz-publish-metrics")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "publishMetrics")
	span.SetAttributes(attribute.String("kubviz-agent", "publish-metrics"))
	defer span.End()

	watchK8sEvents(clientset, js)
	errCh <- nil
}

func publishK8sMetrics(id string, mtype string, mdata *v1.Event, js nats.JetStreamContext, clientset *kubernetes.Clientset) (bool, error) {

	log.Println("*****mdata printing", mdata)

	_, imageNames := getK8sPods(clientset)
	log.Println("***************Image Names:", imageNames)

	ctx := context.Background()
	tracer := otel.Tracer("kubviz-publish-k8smetrics")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "publishK8sMetrics")
	span.SetAttributes(attribute.String("kubviz-agent", "publish-k8smetrics"))
	defer span.End()

	metrics := model.Metrics{
		ID:          id,
		Type:        mtype,
		Event:       mdata,
		ClusterName: ClusterName,
	}
	log.Println("*****struct printing", metrics.Event)

	metricsJson, _ := json.Marshal(metrics)
	_, err := js.Publish(constants.EventSubject, metricsJson)
	if err != nil {
		return true, err
	}
	log.Printf("Metrics with ID:%s has been published\n", id)

	log.Println("*****after marshal printing", mdata)

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

// func getK8sPods(clientset *kubernetes.Clientset) string {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
// 	checkErr(err)
// 	var sb strings.Builder
// 	for i, pod := range pods.Items {
// 		sb.WriteString("Name-" + strconv.Itoa(i) + ": ")
// 		sb.WriteString(pod.Name)
// 		sb.WriteString("   ")
// 		sb.WriteString("Namespace-" + strconv.Itoa(i) + ": ")
// 		sb.WriteString(pod.Namespace)
// 		sb.WriteString("   ")
// 	}
// 	return sb.String()
// }

func getK8sPods(clientset *kubernetes.Clientset) (string, []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	checkErr(err)
	var sb strings.Builder
	var imageNames []string
	for i, pod := range pods.Items {
		sb.WriteString("Name-" + strconv.Itoa(i) + ": ")
		sb.WriteString(pod.Name)
		sb.WriteString("   ")
		sb.WriteString("Namespace-" + strconv.Itoa(i) + ": ")
		sb.WriteString(pod.Namespace)
		sb.WriteString("   ")

		// Check if pod has containers
		if len(pod.Spec.Containers) > 0 {
			for _, container := range pod.Spec.Containers {
				sb.WriteString("Container-" + container.Name + ": ")
				sb.WriteString(container.Image)
				sb.WriteString("   ")
				imageNames = append(imageNames, container.Image)
				log.Println("%%%%")
				log.Println("Pod:", pod.Name, "Container:", container.Name, "Image:", container.Image)
			}
		}
	}
	return sb.String(), imageNames
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

	log.Println("####marshalled events", string(j))

	log.Printf(string(j))
	return string(j)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func LogErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
func watchK8sEvents(clientset *kubernetes.Clientset, js nats.JetStreamContext) {

	ctx := context.Background()
	tracer := otel.Tracer("kubviz-watch-k8sevents")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "watchK8sEvents")
	span.SetAttributes(attribute.String("kubviz-agent", "watch-k8sevents"))
	defer span.End()

	watchlist := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"events",
		v1.NamespaceAll,
		fields.Everything(),
	)
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Event{},
		0, // Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				event := obj.(*v1.Event)
				publishK8sMetrics(string(event.ObjectMeta.UID), "ADD", event, js, clientset)
			},
			DeleteFunc: func(obj interface{}) {
				event := obj.(*v1.Event)
				publishK8sMetrics(string(event.ObjectMeta.UID), "DELETE", event, js, clientset)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				event := newObj.(*v1.Event)
				publishK8sMetrics(string(event.ObjectMeta.UID), "UPDATE", event, js, clientset)
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
func initScheduler(config *rest.Config, js nats.JetStreamContext, cfg config.AgentConfigurations, clientset *kubernetes.Clientset) (s *Scheduler) {
	log := logging.NewLogger()
	s = NewScheduler(log)
	if cfg.OutdatedInterval != "" && cfg.OutdatedInterval != "0" {
		sj, err := NewOutDatedImagesJob(config, js, cfg.OutdatedInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("Outdated", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.GetAllInterval != "" && cfg.GetAllInterval != "0" {
		sj, err := NewKetallJob(config, js, cfg.GetAllInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("GetALL", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.KubeScoreInterval != "" && cfg.KubeScoreInterval != "0" {
		sj, err := NewKubescoreJob(clientset, js, cfg.KubeScoreInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("KubeScore", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.RakkessInterval != "" && cfg.RakkessInterval != "0" {
		sj, err := NewRakkessJob(config, js, cfg.RakkessInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("Rakkess", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.KubePreUpgradeInterval != "" && cfg.KubePreUpgradeInterval != "0" {
		sj, err := NewKubePreUpgradeJob(config, js, cfg.KubePreUpgradeInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("KubePreUpgrade", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	if cfg.TrivyInterval != "" && cfg.TrivyInterval != "0" {
		sj, err := NewTrivyJob(config, js, cfg.TrivyInterval)
		if err != nil {
			log.Fatal("no time interval", err)
		}
		err = s.AddJob("Trivy", sj)
		if err != nil {
			log.Fatal("failed to do job", err)
		}
	}
	return
}
