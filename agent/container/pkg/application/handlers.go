package application

import (
	"context"
	"io"
	"log"

	"net/http"

	"github.com/intelops/kubviz/pkg/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

//githubHandler handles the github webhooks post requests.
func (app *Application) localRegistryHandler(w http.ResponseWriter, r *http.Request) {

	ctx:=context.Background()
	tracer := otel.Tracer("container-gitlab")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "localRegistryHandler")
	span.SetAttributes(attribute.String("http.method", "POST"))
	defer span.End()
	
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
