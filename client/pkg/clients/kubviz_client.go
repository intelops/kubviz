package clients

import (
	"encoding/json"
	"log"

	"github.com/intelops/kubviz/constants"
	"github.com/nats-io/nats.go"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/model"
)

type SubscriptionInfo struct {
	Subject  string
	Consumer string
	Handler  func(msg *nats.Msg)
}

func (n *NATSContext) SubscribeAllKubvizNats(conn clickhouse.DBInterface) {
	subscribe := func(sub SubscriptionInfo) {
		n.stream.Subscribe(sub.Subject, sub.Handler, nats.Durable(sub.Consumer), nats.ManualAck())
	}

	subscriptions := []SubscriptionInfo{
		// {
		// 	Subject:  constants.KetallSubject,
		// 	Consumer: constants.KetallConsumer,
		// 	Handler: func(msg *nats.Msg) {
		// 		msg.Ack()
		// 		var metrics model.Resource
		// 		err := json.Unmarshal(msg.Data, &metrics)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		log.Printf("Ketall Metrics Received: %#v,", metrics)
		// 		conn.InsertKetallEvent(metrics)
		// 		log.Println()
		// 	},
		// },
		// {
		// 	Subject:  constants.RakeesSubject,
		// 	Consumer: constants.RakeesConsumer,
		// 	Handler: func(msg *nats.Msg) {
		// 		msg.Ack()
		// 		var metrics model.RakeesMetrics
		// 		err := json.Unmarshal(msg.Data, &metrics)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		log.Printf("Rakees Metrics Received: %#v,", metrics)
		// 		conn.InsertRakeesMetrics(metrics)
		// 		log.Println()
		// 	},
		// },
		// {
		// 	Subject:  constants.OutdatedSubject,
		// 	Consumer: constants.OutdatedConsumer,
		// 	Handler: func(msg *nats.Msg) {
		// 		msg.Ack()
		// 		var metrics model.CheckResultfinal
		// 		err := json.Unmarshal(msg.Data, &metrics)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		log.Printf("Outdated Metrics Received: %#v,", metrics)
		// 		conn.InsertOutdatedEvent(metrics)
		// 		log.Println()
		// 	},
		// },
		// {
		// 	Subject:  constants.DeprecatedSubject,
		// 	Consumer: constants.DeprecatedConsumer,
		// 	Handler: func(msg *nats.Msg) {
		// 		msg.Ack()
		// 		var metrics model.DeprecatedAPI
		// 		err := json.Unmarshal(msg.Data, &metrics)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		log.Printf("Deprecated API Metrics Received: %#v,", metrics)
		// 		conn.InsertDeprecatedAPI(metrics)
		// 		log.Println()
		// 	},
		// },
		// {
		// 	Subject:  constants.DeletedSubject,
		// 	Consumer: constants.DeletedConsumer,
		// 	Handler: func(msg *nats.Msg) {
		// 		msg.Ack()
		// 		var metrics model.DeletedAPI
		// 		err := json.Unmarshal(msg.Data, &metrics)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		log.Printf("Deleted API Metrics Received: %#v,", metrics)
		// 		conn.InsertDeletedAPI(metrics)
		// 		log.Println()
		// 	},
		// },
		// {
		// 	Subject:  constants.TRIVY_IMAGE_SUBJECT,
		// 	Consumer: constants.Trivy_Image_Consumer,
		// 	Handler: func(msg *nats.Msg) {
		// 		msg.Ack()
		// 		var metrics model.TrivyImage
		// 		err := json.Unmarshal(msg.Data, &metrics)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		log.Printf("Trivy Metrics Received: %#v,", metrics)
		// 		conn.InsertTrivyImageMetrics(metrics)
		// 		log.Println()
		// 	},
		// },
		{
			Subject:  constants.TRIVY_SBOM_SUBJECT,
			Consumer: constants.Trivy_Sbom_Consumer,
			Handler: func(msg *nats.Msg) {
				msg.Ack()
				var metrics model.SbomData
				err := json.Unmarshal(msg.Data, &metrics)
				if err != nil {
					log.Println("failed to unmarshal from nats", err)
					return
				}

				log.Println("sbom log from client side:")
				log.Println("component name :", metrics.ComponentName)
				log.Println("package url :", metrics.PackageUrl)
				log.Println("bom ref :", metrics.BomRef)
				log.Println("serial number :", metrics.SerialNumber)
				log.Println("cyclone dx version :", metrics.CycloneDxVersion)
				log.Println("bom format :", metrics.BomFormat)
				conn.InsertTrivySbomMetrics(metrics)
				log.Println()
			},
		},
		// {
		// 	Subject:  constants.KubvizSubject,
		// 	Consumer: constants.KubvizConsumer,
		// 	Handler: func(msg *nats.Msg) {
		// 		msg.Ack()
		// 		var metrics model.Metrics
		// 		err := json.Unmarshal(msg.Data, &metrics)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		log.Printf("Kubviz Metrics Received: %#v,", metrics)
		// 		conn.InsertKubvizEvent(metrics)
		// 		log.Println()
		// 	},
		// },
		// {
		// 	Subject:  constants.KUBESCORE_SUBJECT,
		// 	Consumer: constants.KubscoreConsumer,
		// 	Handler: func(msg *nats.Msg) {
		// 		msg.Ack()
		// 		var metrics model.KubeScoreRecommendations
		// 		err := json.Unmarshal(msg.Data, &metrics)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		log.Printf("Kubscore Metrics Received: %#v,", metrics)
		// 		conn.InsertKubeScoreMetrics(metrics)
		// 		log.Println()
		// 	},
		// },
		// {
		// 	Subject:  constants.TRIVY_K8S_SUBJECT,
		// 	Consumer: constants.TrivyConsumer,
		// 	Handler: func(msg *nats.Msg) {
		// 		msg.Ack()
		// 		var metrics model.Trivy
		// 		err := json.Unmarshal(msg.Data, &metrics)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 		log.Printf("Trivy Metrics Received: %#v,", metrics)
		// 		conn.InsertTrivyMetrics(metrics)
		// 		log.Println()
		// 	},
		// },
	}

	for _, sub := range subscriptions {
		log.Printf("Creating nats consumer %s with subject: %s \n", sub.Consumer, sub.Subject)
		subscribe(sub)
	}
}
