package otel_test

import (
	"context"
	"testing"

	platformotel "github.com/neuralops/platform/internal/platform/otel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TestW3CTraceContextPropagation(t *testing.T) {
	shutdown, err := platformotel.Init(context.Background(), "gateway-test", "test", "staging")
	if err != nil {
		t.Fatalf("init otel: %v", err)
	}
	defer func() { _ = shutdown(context.Background()) }()

	propagator := otel.GetTextMapPropagator()
	if _, ok := propagator.(propagation.TraceContext); !ok {
		t.Fatalf("expected TraceContext propagator, got %T", propagator)
	}

	traceID, _ := trace.TraceIDFromHex("0102030405060708090a0b0c0d0e0f10")
	spanID, _ := trace.SpanIDFromHex("0102030405060708")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	}))

	carrier := propagation.MapCarrier{}
	propagator.Inject(ctx, carrier)
	if carrier.Get("traceparent") == "" {
		t.Fatal("traceparent header not injected")
	}

	extracted := propagator.Extract(context.Background(), carrier)
	got := trace.SpanContextFromContext(extracted)
	if !got.IsValid() {
		t.Fatal("extracted span context invalid")
	}
	if got.TraceID() != traceID {
		t.Fatalf("trace id mismatch")
	}
	if got.SpanID() != spanID {
		t.Fatalf("span id mismatch")
	}
}
