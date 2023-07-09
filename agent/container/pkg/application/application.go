package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/intelops/kubviz/agent/container/pkg/clients"
	"github.com/intelops/kubviz/agent/container/pkg/config"
	"github.com/intelops/kubviz/agent/container/pkg/handler"
	"github.com/kelseyhightower/envconfig"
)

type Application struct {
	Config     *config.Config
	apiServer  *handler.APIHandler
	conn       *clients.NATSContext
	httpServer *http.Server
}

func New() *Application {
	cfg := &config.Config{}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatalf("Could not parse env Config: %v", err)
	}

	// Connect to NATS
	natsContext, err := clients.NewNATSContext(cfg)
	if err != nil {
		log.Fatal("Error establishing connection to NATS:", err)
	}

	apiServer, err := handler.NewAPIHandler(natsContext)
	if err != nil {
		log.Fatalf("API Handler initialisation failed: %v", err)
	}

	mux := chi.NewMux()
	apiServer.BindRequest(mux)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		Handler: mux,
	}

	return &Application{
		Config:     cfg,
		conn:       natsContext,
		apiServer:  apiServer,
		httpServer: httpServer,
	}
}

func (app *Application) Start() {
	log.Printf("Starting server at %v", app.httpServer.Addr)
	var err error
	if err = app.httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Unexpected server close: %v", err)
	}
	log.Fatalf("Server closed")
}

func (app *Application) Close() {
	log.Printf("Closing the service gracefully")
	app.conn.Close()

	if err := app.httpServer.Shutdown(context.Background()); err != nil {
		log.Printf("Could not close the service gracefully: %v", err)
	}
}
