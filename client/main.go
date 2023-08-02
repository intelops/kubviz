package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/intelops/kubviz/client/pkg/application"
)

func main() {
	log.Println("new client running...")
	app := application.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	app.Close()
}
