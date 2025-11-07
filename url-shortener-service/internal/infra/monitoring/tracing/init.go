package tracing

import (
	"context"
	"os"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
)

// Config holds the configuration options for tracing such as endpoint, sampling ratio, and service metadata
type Config struct {
	ServiceName  string
	Env          string
	Endpoint     string
	SamplerRatio float64
}

// FromEnv loads tracing configuration values from environment variables
func FromEnv() Config {
	r := 1.0
	if v := os.Getenv("OTEL_TRACES_SAMPLER_RATIO"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			r = f
		}
	}
	return Config{
		ServiceName:  os.Getenv("SERVICE_NAME"),
		Env:          os.Getenv("APP_ENV"),
		Endpoint:     os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		SamplerRatio: r,
	}
}

// Init sets up and returns an OpenTelemetry TracerProvider using the provided configuration
func Init(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			attribute.String("environment", cfg.Env),
		),
	)
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	if err != nil {
		tp := trace.NewTracerProvider(trace.WithSampler(trace.TraceIDRatioBased(cfg.SamplerRatio)), trace.WithResource(res))
		otel.SetTracerProvider(tp)
		return tp.Shutdown, err
	}
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
		trace.WithSampler(trace.TraceIDRatioBased(cfg.SamplerRatio)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		),
	)

	return tp.Shutdown, nil
}
