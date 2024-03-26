package handler

import (
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// parse errors
var (
	ErrReadingBody   = errors.New("error reading the request body")
	ErrPublishToNats = errors.New("error while publishing to nats")
)

func (ah *APIHandler) PostEventDockerHub(c *gin.Context) {
	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		tracer := otel.Tracer("dockerhub-container")
		_, span := tracer.Start(c.Request.Context(), "PostEventDockerHub")
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
	log.Printf("Received event from docker artifactory: %v", string(payload))
	err = ah.conn.Publish(payload, "Dockerhub_Registry")
	if err != nil {
		log.Printf("%v: %v", ErrPublishToNats, err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}
