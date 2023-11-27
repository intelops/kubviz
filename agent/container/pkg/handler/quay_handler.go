package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intelops/kubviz/model"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func (ah *APIHandler) PostEventQuayContainer(c *gin.Context) {
	
	tracer := otel.Tracer("quay-container")
	_, span := tracer.Start(c.Request.Context(), "PostEventQuayContainer")
	span.SetAttributes(attribute.String("http.method", "POST"))
	defer span.End()
	
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
	var pushEvent model.QuayImagePushPayload
	err = json.Unmarshal(payload, &pushEvent)
	if err != nil {
		log.Printf("%v: %v", ErrInvalidPayload, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	log.Printf("Received event from Quay Container Registry: %v", pushEvent)

	err = ah.conn.Publish(payload, "Quay_Container_Registry")
	if err != nil {
		log.Printf("%v: %v", ErrPublishToNats, err)
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}