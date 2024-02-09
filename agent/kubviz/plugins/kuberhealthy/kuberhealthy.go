package kuberhealthy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/intelops/kubviz/agent/config"
	"github.com/intelops/kubviz/constants"
	"github.com/intelops/kubviz/model"
	"github.com/intelops/kubviz/pkg/opentelemetry"
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

	var state model.State
	if err := json.Unmarshal(body, &state); err != nil {
		return fmt.Errorf("error unmarshaling response: %w", err)
	}

	return PublishKuberhealthyMetrics(js, state)
}
func boolToUInt8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func errorsToString(errors []string) string {
	return strings.Join(errors, ", ")
}
func PublishKuberhealthyMetrics(js nats.JetStreamContext, state model.State) error {
	ctx := context.Background()
	tracer := otel.Tracer("kuberhealthy")
	_, span := tracer.Start(opentelemetry.BuildContext(ctx), "PublishKuberhealthyMetrics")
	defer span.End()

	for checkName, details := range state.CheckDetails {
		metrics := model.KuberhealthyCheckDetail{
			CurrentUUID:      details.CurrentUUID,
			CheckName:        checkName,
			OK:               boolToUInt8(details.OK),
			Errors:           errorsToString(details.Errors),
			RunDuration:      details.RunDuration,
			Namespace:        details.Namespace,
			Node:             details.Node,
			LastRun:          details.LastRun.Time,
			AuthoritativePod: details.AuthoritativePod,
		}

		metricsJSON, err := json.Marshal(metrics)
		if err != nil {
			log.Printf("Error marshaling metrics of kuberhealthy %s: %v", checkName, err)
			continue
		}

		if _, err := js.Publish(constants.KUBERHEALTHY_SUBJECT, metricsJSON); err != nil {
			log.Printf("Error publishing metrics for kuberhealthy %s: %v", checkName, err)
			continue
		}
	}

	log.Printf("Kuberhealthy metrics have been published")
	return nil
}
