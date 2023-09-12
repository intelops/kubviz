package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
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
	ClusterName  string = os.Getenv("CLUSTER_NAME")
	token        string = os.Getenv("NATS_TOKEN")
	natsurl      string = os.Getenv("NATS_ADDRESS")
	certFilePath string = os.Getenv("CERT_FILE")
	keyFilePath  string = os.Getenv("KEY_FILE")
	caFilePath   string = os.Getenv("CA_FILE")
	//for local testing provide the location of kubeconfig
	// inside the civo file paste your kubeconfig
	// uncomment this line from Dockerfile.Kubviz (COPY --from=builder /workspace/civo /etc/myapp/civo)
	cluster_conf_loc      string = os.Getenv("CONFIG_LOCATION")
	schedulingIntervalStr string = os.Getenv("SCHEDULING_INTERVAL")
)

func runTrivyScans(config *rest.Config, js nats.JetStreamContext) error {
	err := RunTrivyK8sClusterScan(js)
	if err != nil {
		return err
	}
	err = RunTrivyImageScans(config, js)
	if err != nil {
		return err
	}
	err = RunTrivySbomScan(config, js)
	if err != nil {
		return err
	}
	return nil

}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	env := Production
	clusterMetricsChan := make(chan error, 1)
	var (
		config    *rest.Config
		clientset *kubernetes.Clientset
	)
	tlsConfig, err := GetTlsConfig()
	if err != nil {
		log.Println("error while getting tls config ", err)
		time.Sleep(time.Minute * 30)
		log.Fatal("error while getting tls config ", err)
	}
	// connecting with nats ...
	nc, err := nats.Connect(
		natsurl,
		nats.Name("K8s Metrics"),
		nats.Token(token),
		nats.Secure(tlsConfig),
	)

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
		// getK8sEvents(clientset)
		err = runTrivyScans(config, js)
		LogErr(err)
		err = RunKubeScore(clientset, js)
		LogErr(err)
	}

	collectAndPublishMetrics()
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
func LogErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
func watchK8sEvents(clientset *kubernetes.Clientset, js nats.JetStreamContext) {
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
				publishK8sMetrics(string(event.ObjectMeta.UID), "ADD", event, js)
			},
			DeleteFunc: func(obj interface{}) {
				event := obj.(*v1.Event)
				publishK8sMetrics(string(event.ObjectMeta.UID), "DELETE", event, js)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				event := newObj.(*v1.Event)
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

func ReadMtlsCerts(certFile, keyFile, caFile string) (certPEM, keyPEM, caCertPEM []byte, err error) {
	certPEM, err = ReadMTLSFileContents(certFile)
	if err != nil {
		err = fmt.Errorf("error reading cert file: %w", err)
		return
	}

	keyPEM, err = ReadMTLSFileContents(keyFile)
	if err != nil {
		err = fmt.Errorf("error reading key file: %w", err)
		return
	}

	caCertPEM, err = ReadMTLSFileContents(caFile)
	if err != nil {
		err = fmt.Errorf("error reading ca file: %w", err)
		return
	}

	return
}

func OpenMtlsCertFile(path string) (f *os.File, err error) {
	f, err = os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open MTLS cert file: %w", err)
	}
	return f, nil
}

func ReadMTLSFileContents(filePath string) ([]byte, error) {
	file, err := OpenMtlsCertFile(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Error reading file %s: %w", filePath, err)
	}

	return contents, nil
}

func GetTlsConfig() (*tls.Config, error) {
	certPEM, keyPEM, caCertPEM, err := ReadMtlsCerts(certFilePath, keyFilePath, caFilePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read mtls certs %w", err)
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("Error loading X509 key pair from PEM: %w", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertPEM)
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
	}
	return tlsConfig, nil
}
