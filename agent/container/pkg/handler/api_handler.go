package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/intelops/kubviz/agent/container/api"
	"github.com/intelops/kubviz/agent/container/pkg/clients"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type APIHandler struct {
	conn clients.NATSClientInterface
}

const (
	appJSONContentType = "application/json"
	contentType        = "Content-Type"
)

// NewAPIHandler creates a new instance of APIHandler, which is responsible for handling
// various API endpoints related to container events. It takes a NATSContext connection
// as an argument, allowing the handler to interact with a NATS messaging system.
// The returned APIHandler can be used to bind and handle specific routes, such as
// receiving events from Docker Hub or Azure Container Registry.
func NewAPIHandler(conn *clients.NATSContext) (*APIHandler, error) {
	return &APIHandler{
		conn: conn,
	}, nil
}

func (ah *APIHandler) BindRequest(r *gin.Engine) {

	config, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}

	r.Use(otelgin.Middleware(config.ServiceName))

	apiGroup := r.Group("/")
	{
		apiGroup.GET("/api-docs", ah.GetApiDocs)
		apiGroup.GET("/status", ah.GetStatus)
		apiGroup.POST("/event/docker/hub", ah.PostEventDockerHub)
		apiGroup.POST("/event/azure/container", ah.PostEventAzureContainer)
		apiGroup.POST("/event/quay/container", ah.PostEventQuayContainer)
		apiGroup.POST("/event/jfrog/container", ah.PostEventJfrogContainer)
	}
}

// GetApiDocs serves the Swagger API documentation generated from the OpenAPI YAML file.
// It responds with a JSON representation of the API's endpoints, parameters, responses, and other details.
// This endpoint can be used by tools like Swagger UI to provide interactive documentation for the API.
func (ah *APIHandler) GetApiDocs(c *gin.Context) {
	swagger, err := api.GetSwagger()
	fmt.Println(swagger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header(contentType, appJSONContentType)
	c.JSON(http.StatusOK, swagger)
}

// GetStatus responds with the current status of the application. This endpoint can be used
// by monitoring tools to check the health and readiness of the application. It typically
// includes information about the application's state, dependencies, and any ongoing issues.
// In this basic implementation, it simply responds with an OK status, indicating that the
// application is running and ready to handle requests.
func (ah *APIHandler) GetStatus(c *gin.Context) {
	c.Header(contentType, appJSONContentType)
	c.Status(http.StatusOK)
}

var ErrInvalidPayload = errors.New("invalid or malformed Azure Container Registry webhook payload")

// PostEventAzureContainer listens for Azure Container Registry image push events.
// When a new image is pushed, this endpoint receives the event payload, validates it,
// and then publishes it to a NATS messaging system. This allows client of the
// application to subscribe to these events and respond to changes in the container registry.
// If the payload is invalid or the publishing process fails, an error response is returned.
func (ah *APIHandler) PostEventAzureContainer(c *gin.Context) {

	tracer := otel.Tracer("azure-container")
	_, span := tracer.Start(c.Request.Context(), "PostEventAzureContainer")
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

// parse errors
var (
	ErrReadingBody   = errors.New("error reading the request body")
	ErrPublishToNats = errors.New("error while publishing to nats")
)

func (ah *APIHandler) PostEventDockerHub(c *gin.Context) {

	tracer := otel.Tracer("dockerhub-container")
	_, span := tracer.Start(c.Request.Context(), "PostEventDockerHub")
	span.SetAttributes(attribute.String("http.method", "POST"))
	defer span.End()

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("%v: %v", ErrReadingBody, err)
		c.Status(http.StatusBadRequest)
		return
	}
	if len(payload) == 0 || strings.TrimSpace(string(payload)) == "" {
		log.Printf("%v: %v", ErrReadingBody, "empty body")
		c.Status(http.StatusBadRequest)
		return
	}
	log.Printf("Received event from docker artifactory: %v", string(payload))
	err = ah.conn.Publish(payload, "Dockerhub_Registry")
	if err != nil {
		log.Printf("%v: %v", ErrPublishToNats, err)
		c.AbortWithStatus(http.StatusInternalServerError) // Use AbortWithStatus
		return
	}
	c.Status(http.StatusOK)
}

var ErrInvalidPayloads = errors.New("invalid or malformed jfrog Container Registry webhook payload")

func (ah *APIHandler) PostEventJfrogContainer(c *gin.Context) {

	tracer := otel.Tracer("jfrog-container")
	_, span := tracer.Start(c.Request.Context(), "PostEventJfrogContainer")
	span.SetAttributes(attribute.String("http.method", "POST"))
	defer span.End()

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("%v: %v", ErrReadingBody, err)
		c.Status(http.StatusBadRequest)
		return
	}
	if len(payload) == 0 || strings.TrimSpace(string(payload)) == "" {
		log.Printf("%v: %v", ErrReadingBody, "empty body")
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
		c.AbortWithStatus(http.StatusInternalServerError) // Use AbortWithStatus
		return
	}
	c.Status(http.StatusOK)
}

func (ah *APIHandler) PostEventQuayContainer(c *gin.Context) {

	tracer := otel.Tracer("quay-container")
	_, span := tracer.Start(c.Request.Context(), "PostEventQuayContainer")
	span.SetAttributes(attribute.String("http.method", "POST"))
	defer span.End()

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("%v: %v", ErrReadingBody, err)
		c.Status(http.StatusBadRequest)
		return
	}
	if len(payload) == 0 || strings.TrimSpace(string(payload)) == "" {
		log.Printf("%v: %v", ErrReadingBody, "empty body")
		c.Status(http.StatusBadRequest)
		return
	}
	var pushEvent model.QuayImagePushPayload
	err = json.Unmarshal(payload, &pushEvent)
	if err != nil {
		log.Printf("%v: %v", "invalid or malformed Quay Container Registry webhook payload", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	log.Printf("Received event from Quay Container Registry: %v", pushEvent)

	err = ah.conn.Publish(payload, "Quay_Container_Registry")
	if err != nil {
		log.Printf("%v: %v", ErrPublishToNats, err)
		c.AbortWithStatus(http.StatusInternalServerError) // Use AbortWithStatus
		return
	}
	c.Status(http.StatusOK)
}
