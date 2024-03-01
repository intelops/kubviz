package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/intelops/kubviz/agent/git/api"
	"github.com/intelops/kubviz/agent/git/pkg/clients"
	"github.com/intelops/kubviz/agent/git/pkg/config"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/gin-gonic/gin"
)

type Application struct {
	Config *config.Config
	server *http.Server
	conn   *clients.NATSContext
}

func New(conf *config.Config, conn *clients.NATSContext) *Application {
	app := &Application{
		Config: conf,
		conn:   conn,
	}

	app.server = &http.Server{
		// TODO: remove hardcoding
		// Addr:         fmt.Sprintf(":%d", conf.Port),
		Addr:         fmt.Sprintf(":%d", 8081),
		Handler:      app.Routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	return app
}

func (app *Application) Routes() *gin.Engine {
	router := gin.New()

	//opentelemetry
	opentelconfig, err := opentelemetry.GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		router.Use(otelgin.Middleware(opentelconfig.ServiceName))
	}

	api.RegisterHandlers(router, app)
	return router
}

func (app *Application) Start() {
	// TODO: remove hardcoding
	// log.Println("Starting server on port", app.Config.Port)
	log.Printf("Starting server on port %d", 8081)
	if err := app.server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server closed, readon: %v", err)
	}
}

func (app *Application) Close() {
	log.Printf("Closing the service gracefully")
	app.conn.Close()

	if err := app.server.Shutdown(context.Background()); err != nil {
		log.Printf("Could not close the service gracefully: %v", err)
	}
}
