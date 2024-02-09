package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/nats-io/nats.go"

	"context"

	"github.com/intelops/kubviz/pkg/mtlsnats"
	"github.com/intelops/kubviz/pkg/opentelemetry"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/intelops/kubviz/agent/config"
	"github.com/intelops/kubviz/agent/kubviz/plugins/events"
	"github.com/intelops/kubviz/agent/kubviz/plugins/ketall"
	"github.com/intelops/kubviz/agent/kubviz/plugins/kubepreupgrade"

	"github.com/intelops/kubviz/agent/kubviz/plugins/kubescore"
	"github.com/intelops/kubviz/agent/kubviz/plugins/outdated"
	"github.com/intelops/kubviz/agent/kubviz/plugins/rakkess"
	"github.com/intelops/kubviz/agent/kubviz/plugins/trivy"
	"github.com/intelops/kubviz/agent/kubviz/scheduler"

	_ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"

	//  _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
	"github.com/intelops/kubviz/agent/server"
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

	var mtlsConfig mtlsnats.MtlsConfig
	var nc *nats.Conn

	if mtlsConfig.IsEnabled {
		tlsConfig, err := mtlsnats.GetTlsConfig()
		if err != nil {
			log.Println("error while getting tls config ", err)
			time.Sleep(time.Minute * 30)
		} else {
			nc, err = nats.Connect(
				natsurl,
				nats.Name("K8s Metrics"),
				nats.Token(token),
				nats.Secure(tlsConfig),
			)
			if err != nil {
				log.Println("error while connecting with mtls ", err)
			}
		}

	}

	if nc == nil {
		nc, err = nats.Connect(natsurl, nats.Name("K8s Metrics"), nats.Token(token))
		events.CheckErr(err)
	}
	js, err := nc.JetStream()
	events.CheckErr(err)
	err = events.CreateStream(js)
	events.CheckErr(err)
	if env != Production {
		config, err = clientcmd.BuildConfigFromFlags("", cluster_conf_loc)
		if err != nil {
			log.Fatal(err)
		}
		clientset = events.GetK8sClient(config)
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
		clientset = events.GetK8sClient(config)
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

	go events.PublishMetrics(clientset, js, clusterMetricsChan)
	go server.StartServer()
	collectAndPublishMetrics := func() {
		err := outdated.OutDatedImages(config, js)
		events.LogErr(err)
		err = kubepreupgrade.KubePreUpgradeDetector(config, js)
		events.LogErr(err)
		err = ketall.GetAllResources(config, js)
		events.LogErr(err)
		err = rakkess.RakeesOutput(config, js)
		events.LogErr(err)
		err = trivy.RunTrivySbomScan(config, js)
		events.LogErr(err)
		err = trivy.RunTrivyImageScans(config, js)
		events.LogErr(err)
		err = trivy.RunTrivyK8sClusterScan(js)
		events.LogErr(err)
		err = kubescore.RunKubeScore(clientset, js)
		events.LogErr(err)
	}

	collectAndPublishMetrics()

	if cfg.SchedulerEnable { // Assuming "cfg.Schedule" is a boolean indicating whether to schedule or not.
		scheduler := scheduler.InitScheduler(config, js, *cfg, clientset)

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
