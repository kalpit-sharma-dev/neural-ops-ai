package otel

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Init configures OpenTelemetry tracing for a service.
// Set OTEL_EXPORTER_OTLP_ENDPOINT (e.g. http://jaeger:4318) to enable export.
func Init(ctx context.Context, serviceName, version, environment string) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.TraceContext{})

	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if endpoint == "" {
		return func(context.Context) error { return nil }, nil
	}

	opts, err := otlpOptions(endpoint)
	if err != nil {
		return nil, err
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	sampleRate := 1.0
	if strings.EqualFold(environment, "production") {
		sampleRate = 0.1
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sampleRate))),
	)
	otel.SetTracerProvider(provider)

	return provider.Shutdown, nil
}

func otlpOptions(rawEndpoint string) ([]otlptracehttp.Option, error) {
	parsed, err := url.Parse(rawEndpoint)
	if err != nil {
		return nil, fmt.Errorf("parse otlp endpoint: %w", err)
	}

	host := parsed.Host
	if host == "" {
		host = strings.TrimPrefix(strings.TrimPrefix(rawEndpoint, "http://"), "https://")
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(host),
		otlptracehttp.WithURLPath("/v1/traces"),
	}
	if parsed.Scheme != "https" {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	return opts, nil
}
