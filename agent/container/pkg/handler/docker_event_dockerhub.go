package handler

import (
	"errors"
	"io"
	"log"
	"net/http"
)

// parse errors
var (
	ErrReadingBody   = errors.New("error reading the request body")
	ErrPublishToNats = errors.New("error while publishing to nats")
)

func (ah *APIHandler) PostEventDockerHub(w http.ResponseWriter, r *http.Request) {
	defer func() {
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}()
	payload, err := io.ReadAll(r.Body)
	if err != nil || len(payload) == 0 {
		log.Printf("%v: %v", ErrReadingBody, err)
		return
	}
	log.Printf("Received event from docker artifactory: %v", string(payload))
	err = ah.conn.Publish(payload, "Dockerhub_Registry")
	if err != nil {
		log.Printf("%v: %v", ErrPublishToNats, err)
		return
	}
}
