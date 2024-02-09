package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/intelops/kubviz/client/pkg/application"
	"github.com/intelops/kubviz/pkg/opentelemetry"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("new client running...")

	tp, err := opentelemetry.InitTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	
	app := application.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	app.Close()
}
