package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/kube-tarian/kubviz/client/pkg/application"
)

func main() {
	app := application.Start()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	app.Close()
}
