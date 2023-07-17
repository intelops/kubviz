package handler

import (
	"io"
	"log"
	"net/http"
)

func (ah *APIHandler) PostEventDockerHub(w http.ResponseWriter, r *http.Request) {
	event, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Event body read failed: %v", err)
	}

	log.Printf("Received event from docker artifactory: %v", string(event))
	err = ah.conn.Publish(event, "docker registry")
	if err != nil {
		log.Printf("Publish failed for event: %v, reason: %v", string(event), err)
	}
}
