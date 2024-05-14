package nats

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

func (client *Client) CreateStream(streamName string, streamSubjects []string) error {
	js := client.js

	stream, err := js.StreamInfo(streamName)
	if err != nil {
		if err == nats.ErrStreamNotFound {
			client.logger.Printf("Stream does not exist, creating: %s", streamName)
		} else {
			client.logger.Printf("Error getting stream: %s", err)
			return err
		}
	}

	if stream != nil {
		client.logger.Printf("Stream already exists: %s", fmt.Sprintf("%v", stream))
		return nil
	}
	client.logger.Printf("Creating stream %q with subjects %q", streamName, streamSubjects)
	streamInfo, err := js.AddStream(&nats.StreamConfig{
		Name:     streamName,
		Subjects: streamSubjects,
	})

	if err != nil {
		return errors.WithMessage(err, "Error creating stream")
	}
	fmt.Println(streamInfo)
	return nil
}

func (client *Client) Consumer(subject, consumerName string) (interface{}, error) {
	js := client.js
	var data interface{}
	handler := func(msg *nats.Msg) {
		msg.Ack()
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			log.Println("Error unmarshalling message data:", err)
			return
		}
		log.Printf("Data Received: %#v,", data)
	}
	_, err := js.Subscribe(subject, handler, nats.Durable(consumerName), nats.ManualAck())
	if err != nil {
		return nil, fmt.Errorf("error subscribing to stream %s: %w", subject, err)
	}
	return data, nil
}

func (client *Client) Publish(subject string, streamName string, data interface{}) error {
	js := client.js

	resultdata, err := json.Marshal(data)
	if err != nil {
		return errors.WithMessage(err, "Error marshaling data to JSON")
	}
	stream, err := js.StreamInfo(streamName)
	if err != nil {
		if err == nats.ErrStreamNotFound {
			client.logger.Printf("Stream does not exist %s", subject)
		} else {
			client.logger.Printf("Error getting stream: %s", err)
			return err
		}
	}
	if stream == nil {
		return errors.New("Stream does not exist")
	}
	_, err = js.Publish(subject, resultdata)
	if err != nil {
		return errors.WithMessage(err, "Error publishing message")
	}
	return nil
}
