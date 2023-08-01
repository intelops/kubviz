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
	Config       *config.Config
	apiServer    *handler.APIHandler
	conn         *clients.NATSContext
	httpServer   *http.Server
	GithubConfig *config.GithubConfig
}

func New() *Application {
	cfg := &config.Config{}
	if err := envconfig.Process("", cfg); err != nil {
		log.Fatalf("Could not parse nats env Config: %v", err)
	}
	githubcfg := &config.GithubConfig{}
	if err := envconfig.Process("", githubcfg); err != nil {
		log.Fatalf("Could not parse github env Config: %v", err)
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
	chi.Walk(mux, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		fmt.Printf("[%s]: '%s' has %d middlewares.\n", method, route, len(middlewares))
		return nil
	})
	httpServer := &http.Server{
		// TODO: remove hardcoding
		// Addr:    fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		Addr:    fmt.Sprintf(":%d", 8082),
		Handler: mux,
	}

	return &Application{
		Config:       cfg,
		conn:         natsContext,
		apiServer:    apiServer,
		httpServer:   httpServer,
		GithubConfig: githubcfg,
	}
}

func (app *Application) Start() {
	// TODO: remove hard coding
	// log.Printf("Starting server at %v", app.httpServer.Addr)
	log.Printf("Starting server at %v", 8082)
	var err error
	app.GithubContainerWatch()
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
