package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/intelops/kubviz/agent/container/api"
	"github.com/intelops/kubviz/agent/container/pkg/clients"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type APIHandler struct {
	conn *clients.NATSContext
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
