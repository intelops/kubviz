package opentelemetry

import (
	"context"
	"log"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Configurations struct {
	ServiceName  string `envconfig:"APPLICATION_NAME" default:"Kubviz"`
	CollectorURL string `envconfig:"OPTEL_URL" default:"otelcollector.azureagent.optimizor.app:80"`
}

func GetConfigurations() (opteConfig *Configurations, err error) {
	opteConfig = &Configurations{}
	if err = envconfig.Process("", opteConfig); err != nil {
		return nil, errors.WithStack(err)
	}
	return
}

func InitTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

    config, err := GetConfigurations()
	if err != nil {
		log.Println("Unable to read open telemetry configurations")
		return nil, err
	}

    headers := map[string]string{
		"signoz-service-name": config.ServiceName,
	}

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(config.CollectorURL),
        otlptracegrpc.WithHeaders(headers),
		otlptracegrpc.WithInsecure(),
	)

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Fatalf("failed to initialize exporter: %e", err)
	}

	res, err := resource.New(
        ctx,
        resource.WithAttributes(
            attribute.String("service.name", config.ServiceName),
			attribute.String("library.language", "go"),

        ),
    )
	if err != nil {
		log.Fatalf("failed to initialize resource: %e", err)
	}

	// Create the trace provider
	tp := sdktrace.NewTracerProvider(
        trace.WithSampler(trace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set the global trace provider
	otel.SetTracerProvider(tp)

	// Set the propagator
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(propagator)

	return tp, nil
}

func BuildContext(ctx context.Context) context.Context {
	newCtx, _ := context.WithCancel(ctx)
	return newCtx
}