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
	"github.com/intelops/kubviz/pkg/nats/sdk"
	"github.com/intelops/kubviz/pkg/opentelemetry"
	"github.com/kuberhealthy/kuberhealthy/v2/pkg/health"

	//"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
)

func StartKuberHealthy(natsCli *sdk.NATSClient) {
	khConfig, err := config.GetKuberHealthyConfig()
	if err != nil {
		log.Fatalf("Error getting Kuberhealthy config: %v", err)
	}

	ticker := time.NewTicker(khConfig.PollInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := pollAndPublishKuberhealthy(khConfig.KuberhealthyURL, natsCli); err != nil {
			log.Printf("Error polling and publishing Kuberhealthy metrics: %v", err)
		}
	}
}
func pollAndPublishKuberhealthy(url string, natsCli *sdk.NATSClient) error {
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

	return PublishKuberhealthyMetrics(natsCli, state)
}

func PublishKuberhealthyMetrics(natsCli *sdk.NATSClient, state health.State) error {
	ctx := context.Background()
	tracer := otel.Tracer("kuberhealthy")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "PublishKuberhealthyMetrics")
	defer span.End()

	metricsJSON, err := json.Marshal(state)
	if err != nil {
		log.Printf("Error marshaling metrics of kuberhealthy %v", err)
		return err
	}
	if err := natsCli.Publish(constants.KUBERHEALTHY_SUBJECT, metricsJSON); err != nil {
		log.Printf("Error publishing metrics for kuberhealthy %v", err)
		return err
	}
	log.Printf("Kuberhealthy metrics have been published")
	return nil
}
