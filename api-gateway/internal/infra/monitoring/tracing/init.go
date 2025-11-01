package tracing

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
)

// Config defines tracing configuration
type Config struct {
	ServiceName string
	Env         string
	Ratio       float64
	Endpoint    string
}

// FromEnv loads config from environment variables
func FromEnv() Config {
	return Config{
		ServiceName: os.Getenv("SERVICE_NAME"),
		Env:         os.Getenv("APP_ENV"),
		Ratio:       1.0,
		Endpoint:    os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}
}

// Init initializes OpenTelemetry TracerProvider (stdout exporter for local)
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "jaeger:4317" // default for local Jaeger container
	}

	// 1. Create exporter: print traces to stdout in readable format
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		return nil, err
	}

	// 2. Create resource to describe this service
	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Env),
		),
	)
	if err != nil {
		return nil, err
	}

	// 3. Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.Ratio))),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	// 4. Register as global provider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	log.Printf("[Tracing] Initialized tracer for service=%s env=%s\n", cfg.ServiceName, cfg.Env)

	return tp.Shutdown, nil
}
