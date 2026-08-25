// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package otel

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/log/noop"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

type Otel struct {
	Log    Logger
	Metric Meter
	Trace  Tracer
}

func (o *Otel) IsLogAvailable() bool {
	return o.Log != nil
}

func (o *Otel) Clone() Otel {
	return Otel{
		Log:    o.Log,
		Metric: o.Metric,
		Trace:  o.Trace,
	}
}

type SetupOption struct {
	EnableMetric   bool
	EnableTrace    bool
	EnableLog      bool
	ServiceName    string
	ServiceVersion string
}

func CobraFuncEWithLogger(ctx context.Context, name string, logLevel string, callback func(otl Otel) error) error {
	otl := NewOtel(name, logLevel)
	if err := callback(otl); err != nil {
		otl.Log.Errorf(ctx, "executiing %s error bacause %s", name, err)
		return err
	}

	return nil
}

func NewOtel(name string, logLevel string) Otel {
	return Otel{
		Log:    NewLogger(name, logLevel),
		Metric: NewOtelMeter(name),
		Trace:  NewOtelTracer(name),
	}
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context, opt SetupOption) (shutdown func(context.Context) error, err error) {
	r := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(opt.ServiceName),
		semconv.DeviceID(uuid.NewString()),
	)
	if err != nil {
		return nil, err
	}

	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	if !opt.EnableTrace {
		otel.SetTracerProvider(tracenoop.NewTracerProvider())
	} else {
		tracerProvider, xerr := newTraceProvider(r)
		if xerr != nil {
			handleErr(xerr)
			return
		}
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)

	}

	if !opt.EnableMetric {
		otel.SetMeterProvider(metricnoop.NewMeterProvider())
	} else {
		// Set up meter provider.
		meterProvider, xerr := newMeterProvider(r)
		if xerr != nil {
			handleErr(xerr)
			return
		}
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)

	}

	if !opt.EnableLog {
		global.SetLoggerProvider(noop.NewLoggerProvider())
	} else {
		loggerProvider, xerr := newLoggerProvider(r)
		if xerr != nil {
			handleErr(xerr)
			return
		}
		shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
		global.SetLoggerProvider(loggerProvider)
	}

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(r *resource.Resource) (*trace.TracerProvider, error) {
	httpExporter, err := otlptracehttp.New(context.TODO())
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		// trace.WithBatcher(traceExporter, trace.WithBatchTimeout(time.Second)),
		trace.WithBatcher(httpExporter, trace.WithBatchTimeout(time.Second)),
		trace.WithResource(r),
	)

	return traceProvider, nil
}

func newMeterProvider(r *resource.Resource) (*metric.MeterProvider, error) {
	// metricExporter, err := stdoutmetric.New()
	// if err != nil {
	// 	return nil, err
	// }

	metricHttp, err := otlpmetrichttp.New(context.TODO())
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		// metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(3*time.Second))),
		metric.WithReader(metric.NewPeriodicReader(metricHttp, metric.WithInterval(3*time.Second))),
		metric.WithResource(r),
	)
	return meterProvider, nil
}

func newLoggerProvider(r *resource.Resource) (*log.LoggerProvider, error) {
	httpExporter, err := otlploghttp.New(context.TODO())
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(httpExporter)),
		log.WithResource(r),
	)

	return loggerProvider, nil
}
