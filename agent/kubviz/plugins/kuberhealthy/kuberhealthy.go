package kuberhealthy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/intelops/kubviz/agent/config"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/kuberhealthy/kuberhealthy/v2/pkg/health"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
)

func StartKuberHealthy(js nats.JetStreamContext) {
	khConfig, err := config.GetKuberHealthyConfig()
	if err != nil {
		log.Fatalf("Error getting Kuberhealthy config: %v", err)
	}

	ticker := time.NewTicker(khConfig.PollInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := pollAndPublishKuberhealthy(khConfig.KuberhealthyURL, js); err != nil {
			log.Printf("Error polling and publishing Kuberhealthy metrics: %v", err)
		}
	}
}
func pollAndPublishKuberhealthy(url string, js nats.JetStreamContext) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making GET request to Kuberhealthy: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	var state health.State
	if err := json.Unmarshal(body, &state); err != nil {
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	return PublishKuberhealthyMetrics(js, state)
}

func PublishKuberhealthyMetrics(js nats.JetStreamContext, state health.State) error {
	// opentelemetry
	opentelconfig, errs := opentelemetry.GetConfigurations()
	if errs != nil {
		log.Println("Unable to read open telemetry configurations")
	}
	if opentelconfig.IsEnabled {
		ctx := context.Background()
		tracer := otel.Tracer("kuberhealthy")
		_, span := tracer.Start(opentelemetry.BuildContext(ctx), "PublishKuberhealthyMetrics")
		defer span.End()
	}

	metricsJSON, err := json.Marshal(state)
	if err != nil {
		log.Printf("Error marshaling metrics of kuberhealthy %v", err)
		return err
	}

	if _, err := js.Publish(constants.KUBERHEALTHY_SUBJECT, metricsJSON); err != nil {
		log.Printf("Error publishing metrics for kuberhealthy %v", err)
		return err
	}

	log.Printf("Kuberhealthy metrics have been published")
	return nil
}
