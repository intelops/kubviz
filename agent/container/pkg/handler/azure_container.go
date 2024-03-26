package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var ErrInvalidPayload = errors.New("invalid or malformed Azure Container Registry webhook payload")

// PostEventAzureContainer listens for Azure Container Registry image push events.
// When a new image is pushed, this endpoint receives the event payload, validates it,
// and then publishes it to a NATS messaging system. This allows client of the
// application to subscribe to these events and respond to changes in the container registry.
// If the payload is invalid or the publishing process fails, an error response is returned.
func (ah *APIHandler) PostEventAzureContainer(c *gin.Context) {

	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		tracer := otel.Tracer("azure-container")
		_, span := tracer.Start(c.Request.Context(), "PostEventAzureContainer")
		span.SetAttributes(attribute.String("http.method", "POST"))
		defer span.End()
	}

	defer func() {
		_, _ = io.Copy(io.Discard, c.Request.Body)
		_ = c.Request.Body.Close()
	}()
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil || len(payload) == 0 {
		log.Printf("%v: %v", ErrReadingBody, err)
		c.Status(http.StatusBadRequest)
		return
	}

	var pushEvent model.AzureContainerPushEventPayload
	err = json.Unmarshal(payload, &pushEvent)
	if err != nil {
		log.Printf("%v: %v", ErrInvalidPayload, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	log.Printf("Received event from Azure Container Registry: %v", pushEvent)

	err = ah.conn.Publish(payload, "Azure_Container_Registry")
	if err != nil {
		log.Printf("%v: %v", ErrPublishToNats, err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}
