package otel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Tracer interface {
	trace.Tracer
}

type Meter interface {
	metric.Meter
}

func NewOtelTracer(name string) Tracer {
	return otel.Tracer(name)
}

func NewOtelMeter(name string) Meter {
	return otel.Meter(name)
}
