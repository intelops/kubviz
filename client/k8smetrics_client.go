package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/kube-tarian/kubviz/clickhouse"
	"github.com/kube-tarian/kubviz/model"
	"github.com/nats-io/nats.go"
	"gopkg.in/yaml.v2"
)

// to read the token from env variables

var (
	token    string = os.Getenv("NATS_TOKEN")
	natsurl  string = os.Getenv("NATS_ADDRESS")
	dbAdress string = os.Getenv("DB_ADDRESS")
	dbPort   string = os.Getenv("DB_PORT")
	url      string = fmt.Sprintf("tcp://%s:%s?debug=true", dbAdress, dbPort)
)

func main() {

	// Connect to NATS

	nc, err := nats.Connect(natsurl, nats.Name("K8s Metrics"), nats.Token(token))
	checkErr(err)
	// log.Println(nc)
	js, err := nc.JetStream()
	log.Print("jetstream context:", js)
	checkErr(err)

	stream, err := js.StreamInfo("METRICS")
	checkErr(err)
	log.Println("Stream Info:", stream)
	//Get clickhouse connection
	connection, err := clickhouse.GetClickHouseConnection(url)
	if err != nil {
		log.Fatal(err)
	}

	//Create schema
	clickhouse.CreateSchema(connection)
	clickhouse.CreateKubePugSchema(connection)
	//Get db data
	// data, err := clickhouse.RetrieveEvent(connection)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Printf("DB: %s", data)

	// datas, err := clickhouse.RetriveKubepugEvent(connection)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// log.Printf("DB: %#v", datas)

	// Create durable consumer monitor
	// consume kubepug result and insert in clickhouse
	js.Subscribe("METRICS.kubepug", func(msg *nats.Msg) {
		msg.Ack()
		var metrics model.Result
		err := json.Unmarshal(msg.Data, &metrics)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Kubepug Metrics Received: %#v", metrics)
		clickhouse.InsertKubepugEvent(connection, metrics)
		log.Println()
	}, nats.Durable("KUBEPUG_EVENTS_CONSUMER"), nats.ManualAck())

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
		log.Printf("Metrics received - subject: %s, ID: %s, Type: %s, Event: %s, ClusterName %s\n", msg.Subject, metrics.ID, metrics.Type, y, metrics.ClusterName)
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
