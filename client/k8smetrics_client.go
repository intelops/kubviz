package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"runtime"

	"github.com/ghodss/yaml"
	"github.com/kube-tarian/kubviz/clickhouse"
	"github.com/kube-tarian/kubviz/model"
	"github.com/nats-io/nats.go"
)

func main() {

	// Connect to NATS
	nc, err := nats.Connect("127.0.0.1", nats.Name("K8s Metrics"), nats.Token("UfmrJOYwYCCsgQvxvcfJ3BdI6c8WBbnD"))
	checkErr(err)
	log.Println(nc)
	js, err := nc.JetStream()
	log.Print(js)
	checkErr(err)

	stream, err := js.StreamInfo("METRICS")
	checkErr(err)
	log.Println(stream)
	//Get clickhouse connection
	connection, err := clickhouse.GetClickHouseConnection()
	if err != nil {
		log.Fatal(err)
	}

	//Create schema
	clickhouse.CreateSchema(connection)

	//Get db data
	// data, err := clickhouse.RetrieveEvent(connection)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Printf("DB: %s", data)

	// Create durable consumer monitor
	js.Subscribe("METRICS.event", func(msg *nats.Msg) {
		msg.Ack()
		var metrics model.Metrics
		err := json.Unmarshal(msg.Data, &metrics)
		if err != nil {
			log.Fatal(err)
		}
		y, err := yaml.Marshal(metrics.Event)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		//fmt.Printf("Add event: %s \n", y)
		log.Printf("Metrics received - subject: %s, ID: %s, Type: %s, Event: %s\n", msg.Subject, metrics.ID, metrics.Type, y)
		// Insert event
		clickhouse.InsertEvent(connection, metrics)
		log.Println()
	}, nats.Durable("EVENTS_CONSUMER"), nats.ManualAck())

	runtime.Goexit()
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func jsonPrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}
