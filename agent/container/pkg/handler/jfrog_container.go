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

var ErrInvalidPayloads = errors.New("invalid or malformed jfrog Container Registry webhook payload")

func (ah *APIHandler) PostEventJfrogContainer(c *gin.Context) {

	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		tracer := otel.Tracer("jfrog-container")
		_, span := tracer.Start(c.Request.Context(), "PostEventJfrogContainer")
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

	var pushEvent model.JfrogContainerPushEventPayload
	err = json.Unmarshal(payload, &pushEvent)
	if err != nil {
		log.Printf("%v: %v", ErrInvalidPayloads, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	log.Printf("Received event from jfrog Container Registry: %v", pushEvent)

	err = ah.conn.Publish(payload, "Jfrog_Container_Registry")
	if err != nil {
		log.Printf("%v: %v", ErrPublishToNats, err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}
