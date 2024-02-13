package clients

import (
	"context"
	"encoding/json"
	"log"

	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/kelseyhightower/envconfig"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/client/pkg/config"
	"github.com/intelops/kubviz/model"
)

type SubscriptionInfo struct {
	Subject  string
	Consumer string
	Handler  func(msg *nats.Msg)
}

func (n *NATSContext) SubscribeAllKubvizNats(conn clickhouse.DBInterface) {

	ctx := context.Background()
	tracer := otel.Tracer("kubviz-client")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "SubscribeAllKubvizNats")
	span.SetAttributes(attribute.String("kubviz-subscribe", "subscribe"))
	defer span.End()
	cfg := &config.Config{}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatalf("Could not parse env Config: %v", err)
	}
	subscribe := func(sub SubscriptionInfo) {
		n.stream.Subscribe(sub.Subject, sub.Handler, nats.Durable(sub.Consumer), nats.ManualAck())
	}

	subscriptions := []SubscriptionInfo{
		{
			Subject:  constants.KetallSubject,
			Consumer: cfg.KetallConsumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.Resource
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Ketall Metrics Received: %#v,", metrics)
				conn.InsertKetallEvent(metrics)
				log.Println()
			},
		},
		{
			Subject:  constants.RakeesSubject,
			Consumer: cfg.RakeesConsumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.RakeesMetrics
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Rakees Metrics Received: %#v,", metrics)
				conn.InsertRakeesMetrics(metrics)
				log.Println()
			},
		},
		{
			Subject:  constants.OutdatedSubject,
			Consumer: cfg.OutdatedConsumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.CheckResultfinal
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Outdated Metrics Received: %#v,", metrics)
				conn.InsertOutdatedEvent(metrics)
				log.Println()
			},
		},
		{
			Subject:  constants.DeprecatedSubject,
			Consumer: cfg.DeprecatedConsumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.DeprecatedAPI
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Deprecated API Metrics Received: %#v,", metrics)
				conn.InsertDeprecatedAPI(metrics)
				log.Println()
			},
		},
		{
			Subject:  constants.DeletedSubject,
			Consumer: cfg.DeletedConsumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.DeletedAPI
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Deleted API Metrics Received: %#v,", metrics)
				conn.InsertDeletedAPI(metrics)
				log.Println()
			},
		},
		{
			Subject:  constants.TRIVY_IMAGE_SUBJECT,
			Consumer: cfg.TrivyImageConsumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.TrivyImage
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Trivy Metrics Received: %#v,", metrics)
				conn.InsertTrivyImageMetrics(metrics)
				log.Println()
			},
		},
		{
			Subject:  constants.TRIVY_SBOM_SUBJECT,
			Consumer: cfg.TrivySbomConsumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.SbomData
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Println("failed to unmarshal from nats", err)
					return
				}
				log.Printf("Trivy sbom Metrics Received: %#v,", metrics)
				conn.InsertTrivySbomMetrics(metrics)
				log.Println()
			},
		},
		{
			Subject:  constants.KUBERHEALTHY_SUBJECT,
			Consumer: cfg.KuberhealthyConsumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.KuberhealthyCheckDetail
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Println("failed to unmarshal from nats", err)
					return
				}
				log.Printf("Kuberhealthy Metrics Received: %#v,", metrics)
				conn.InsertKuberhealthyMetrics(metrics)
				log.Println()
			},
		},
		{
			Subject:  constants.KubvizSubject,
			Consumer: cfg.KubvizConsumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.Metrics
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Kubviz Metrics Received: %#v,", metrics)
				conn.InsertKubvizEvent(metrics)
				log.Println()
			},
		},
		{
			Subject:  constants.KUBESCORE_SUBJECT,
			Consumer: cfg.KubscoreConsumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.KubeScoreRecommendations
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Kubscore Metrics Received: %#v,", metrics)
				conn.InsertKubeScoreMetrics(metrics)
				log.Println()
			},
		},
		{
			Subject:  constants.TRIVY_K8S_SUBJECT,
			Consumer: cfg.TrivyConsumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.Trivy
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("Trivy Metrics Received: %#v,", metrics)
				conn.InsertTrivyMetrics(metrics)
				log.Println()
			},
		},
	}

	for _, sub := range subscriptions {
		log.Printf("Creating nats consumer %s with subject: %s \n", sub.Consumer, sub.Subject)
		subscribe(sub)
	}
}
