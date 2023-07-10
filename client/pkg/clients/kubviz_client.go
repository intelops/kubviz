package clients

import (
	"encoding/json"
	"log"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/model"
	"github.com/nats-io/nats.go"
)

const (
	ketallSubject      = "METRICS.ketall"
	ketallConsumer     = "KETALL_EVENTS_CONSUMER"
	rakeesSubject      = "METRICS.rakees"
	rakeesConsumer     = "RAKEES_METRICS_CONSUMER"
	outdatedSubject    = "METRICS.outdated"
	outdatedConsumer   = "OUTDATED_EVENTS_CONSUMER"
	deprecatedSubject  = "METRICS.deprecatedAPI"
	deprecatedConsumer = "DEPRECATED_API_CONSUMER"
	deletedSubject     = "METRICS.deletedAPI"
	deletedConsumer    = "DELETED_API_CONSUMER"
	kubvizSubject      = "METRICS.kubvizevent"
	kubvizConsumer     = "KUBVIZ_EVENTS_CONSUMER"
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
		{
			Subject:  ketallSubject,
			Consumer: ketallConsumer,
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
			Subject:  rakeesSubject,
			Consumer: rakeesConsumer,
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
			Subject:  outdatedSubject,
			Consumer: outdatedConsumer,
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
			Subject:  deprecatedSubject,
			Consumer: deprecatedConsumer,
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
			Subject:  deletedSubject,
			Consumer: deletedConsumer,
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
			Subject:  kubvizSubject,
			Consumer: kubvizConsumer,
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
	}

	for _, sub := range subscriptions {
		subscribe(sub)
	}
}
