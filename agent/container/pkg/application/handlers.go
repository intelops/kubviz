package application

import (
	"io"
	"log"

	"net/http"
)

// githubHandler handles the github webhooks post requests.
func (app *Application) localRegistryHandler(w http.ResponseWriter, r *http.Request) {

	event, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Event body read failed: %v", err)
	}
	log.Printf("Received event from gitlab: %v", string(event))
	err = app.conn.Publish(event, "gitlab")
	if err != nil {
		log.Printf("Publish failed for event: %v, reason: %v", string(event), err)
	}
}
