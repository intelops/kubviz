package clients

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/intelops/kubviz/client/pkg/clickhouse"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var (
	ErrUnmarshalBuildPayload = errors.New("error while unmarshal the dockerhub build payload")
)

type Container string

// constant variables to use with nats stream and
// nats publishing
const (
	containerSubjects Container = "CONTAINERMETRICS.*"
	containerSubject  Container = "CONTAINERMETRICS.git"
	containerConsumer Container = "CONTAINER_EVENT_CONSUMER"
)

func (n *NATSContext) SubscribeContainerNats(conn clickhouse.DBInterface) {

	ctx := context.Background()
	tracer := otel.Tracer("container-client")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "SubscribeContainerNats")
	span.SetAttributes(attribute.String("container-subscribe", "Subscribe"))
	defer span.End()

	n.stream.Subscribe(string(containerSubject), func(msg *nats.Msg) {
		msg.Ack()
		repoName := msg.Header.Get("REPO_NAME")
		if repoName == "Dockerhub_Registry" {
			var pl model.BuildPayload
			err := json.Unmarshal(msg.Data, &pl)
			if err != nil {
				log.Printf("%v", ErrUnmarshalBuildPayload)
				return
			}
			var hub model.DockerHubBuild
			t := time.Unix(int64(pl.Repository.DateCreated), 0)
			hub.DateCreated = t.Format("2006-01-02 15:04:05")
			hub.PushedBy = pl.PushData.Pusher
			hub.ImageTag = pl.PushData.Tag
			hub.RepositoryName = pl.Repository.Name
			hub.Owner = pl.Repository.Owner
			hub.Event = string(msg.Data)
			conn.InsertContainerEventDockerHub(hub)
			log.Println("Inserted DockerHub Container metrics:", string(msg.Data))
		} else if repoName == "Github_Registry" {
			conn.InsertContainerEventGithub(string(msg.Data))
			log.Println("Inserted Github Container metrics:", string(msg.Data))
		} else if repoName == "Azure_Container_Registry" {
			var pushEvent model.AzureContainerPushEventPayload
			err := json.Unmarshal(msg.Data, &pushEvent)
			if err != nil {
				log.Printf("Error while unmarshaling Azure Container Registry payload: %v", err)
				return
			}
			// Extract the necessary information from pushEvent and insert into ClickHouse
			conn.InsertContainerEventAzure(pushEvent)
			log.Println("Inserted Azure Container Registry metrics:", string(msg.Data))
		} else if repoName == "Quay_Container_Registry" {
			var pushEvent model.QuayImagePushPayload
			err := json.Unmarshal(msg.Data, &pushEvent)
			if err != nil {
				log.Printf("Error while unmarshaling Quay Container Registry payload: %v", err)
				return
			}
			// Extract the necessary information from pushEvent and insert into ClickHouse
			conn.InsertContainerEventQuay(pushEvent)
			log.Println("Inserted Quay Container Registry metrics:", string(msg.Data))
		} else if repoName == "Jfrog_Container_Registry" {
			var pushEvent model.JfrogContainerPushEventPayload
			err := json.Unmarshal(msg.Data, &pushEvent)
			if err != nil {
				log.Printf("Error while unmarshaling Jfrog Container Registry payload: %v", err)
				return
			}
			// Extract the necessary information from pushEvent and insert into ClickHouse.
			conn.InsertContainerEventJfrog(pushEvent)
			log.Println("Inserted Jfrog Container Registry metrics:", string(msg.Data))
		}

	}, nats.Durable(string(containerConsumer)), nats.ManualAck())
}
