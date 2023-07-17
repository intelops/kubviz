package handler

import (
	"io"
	"log"
	"net/http"
)

func (ah *APIHandler) PostEventDockerGithub(w http.ResponseWriter, r *http.Request) {
	event, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Event body read failed: %v", err)
	}

	log.Printf("Received docker event from github artifactory: %v", string(event))
	err = ah.conn.Publish(event, "Github_Registory")
	if err != nil {
		log.Printf("Publish failed for event: %v, reason: %v", string(event), err)
	}
}
