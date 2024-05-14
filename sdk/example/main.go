package main

import (
	"fmt"
	"log"
	"time"

	"github.com/intelops/kubviz/sdk/pkg/clickhouse"
	"github.com/intelops/kubviz/sdk/pkg/nats"
	"github.com/intelops/kubviz/sdk/pkg/sdk"
)

func main() {
	natsConfig, err := nats.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load NATS config: %v", err)
	}

	chConfig, err := clickhouse.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load ClickHouse config: %v", err)
	}

	mySDK, err := sdk.New(natsConfig, chConfig)
	if err != nil {
		log.Fatalf("Failed to initialize SDK: %v", err)
	}
	streamName := "Simple"
	streamSubjects := "Simple.*"
	err = mySDK.CreateNatsStream(streamName, []string{streamSubjects})
	if err != nil {
		fmt.Println("Error creating NATS Stream:", err)
		return
	}

	time.Sleep(2 * time.Second)

	data := map[string]interface{}{
		"key":   "value",
		"count": 42,
	}
	subject := "Simple.event"
	err = mySDK.PublishToNats(subject, streamName, data)
	if err != nil {
		fmt.Println("Error publishing message to NATS:", err)
		return
	}
	time.Sleep(2 * time.Second)
	consumerName := "myConsumer"
	err = mySDK.ConsumeNatsData(subject, consumerName)
	if err != nil {
		fmt.Println("Error creating NATS consumer:", err)
		return
	}
	err = mySDK.ClickHouseInsertData("mytable", data)
	if err != nil {
		fmt.Println("Error while inserting data into nats:", err)
		return
	}
}
