package clients

import (
	"encoding/json"
	"log"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/nats-io/nats.go"
)

type Container string

// constant variables to use with nats stream and
// nats publishing
const (
	containerSubjects Container = "CONTAINERMETRICS.*"
	containerSubject  Container = "CONTAINERMETRICS.git"
	containerConsumer Container = "container-event-consumer"
)

// func (n *NATSContext) SubscribeContainerNats(conn clickhouse.DBInterface) {
// 	n.stream.Subscribe(string(containerSubject), func(msg *nats.Msg) {
// 		type events struct {
// 			Events []json.RawMessage `json:"events"`
// 		}

// 		eventDocker := &events{}
// 		err := json.Unmarshal(msg.Data, &eventDocker)
// 		if err == nil {
// 			log.Println(eventDocker)
// 			msg.Ack()
// 			repoName := msg.Header.Get("REPO_NAME")
// 			type newEvent struct {
// 				RepoName string          `json:"repoName"`
// 				Event    json.RawMessage `json:"event"`
// 			}

// 			for _, event := range eventDocker.Events {
// 				event := &newEvent{
// 					RepoName: repoName,
// 					Event:    event,
// 				}

// 				eventsJSON, err := json.Marshal(event)
// 				if err != nil {
// 					log.Printf("Failed to marshall with repo name going ahead with only event, %v", err)
// 					eventsJSON = msg.Data
// 				}
// 				conn.InsertContainerEvent(string(eventsJSON))
// 			}
// 		} else {
// 			log.Printf("Failed to unmarshal event, %v", err)
// 			conn.InsertContainerEvent(string(msg.Data))
// 		}

//			log.Println("Inserted metrics:", string(msg.Data))
//		}, nats.Durable(string(containerConsumer)), nats.ManualAck())
//	}
func (n *NATSContext) SubscribeContainerNats(conn clickhouse.DBInterface) {
	n.stream.Subscribe(string(containerSubject), func(msg *nats.Msg) {
		type pubData struct {
			Metrics json.RawMessage `json:"event"`
			Repo    string          `json:"repoName"`
		}
		msg.Ack()
		repoName := msg.Header.Get("REPO_NAME")
		metrics := &pubData{
			Metrics: json.RawMessage(msg.Data),
			Repo:    repoName,
		}
		data, err := json.Marshal(metrics)
		if err != nil {
			log.Fatal(err)
		}
		conn.InsertContainerEvent(string(data))
		log.Println("Inserted Container metrics:", string(msg.Data))
	}, nats.Durable(string(containerConsumer)), nats.ManualAck())
}
