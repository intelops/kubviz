package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/intelops/kubviz/agent/container/pkg/application"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	app := application.New()
	go app.GithubContainerWatch()
	go app.Start()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	app.Close()
}
