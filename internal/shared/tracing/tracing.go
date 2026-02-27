package tracing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"p2p_wallet/internal/shared/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

func Init(ctx context.Context, cfg config.TracingConfig, serviceName, serviceVersion, environment string) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	)

	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(serviceVersion),
			semconv.DeploymentEnvironmentName(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create tracing resource: %w", err)
	}

	exp, err := newExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithSampler(newSampler(cfg.Sampler, cfg.SamplerArg)),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

func newExporter(ctx context.Context, cfg config.TracingConfig) (sdktrace.SpanExporter, error) {
	protocol := strings.ToLower(strings.TrimSpace(cfg.Protocol))
	switch protocol {
	case "", "grpc":
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(cfg.Endpoint),
			otlptracegrpc.WithTimeout(5 * time.Second),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}

		exp, err := otlptracegrpc.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("create otlp/grpc trace exporter: %w", err)
		}
		return exp, nil
	case "http", "http/protobuf":
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.Endpoint),
			otlptracehttp.WithTimeout(5 * time.Second),
		}
		if cfg.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}

		exp, err := otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("create otlp/http trace exporter: %w", err)
		}
		return exp, nil
	default:
		return nil, fmt.Errorf("unsupported OTLP protocol %q", cfg.Protocol)
	}
}

func newSampler(raw string, ratio float64) sdktrace.Sampler {
	base := sdktrace.TraceIDRatioBased(ratio)

	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "always_on", "alwayson":
		return sdktrace.AlwaysSample()
	case "always_off", "alwaysoff":
		return sdktrace.NeverSample()
	case "traceidratio":
		return base
	case "", "parentbased_traceidratio":
		return sdktrace.ParentBased(base)
	default:
		return sdktrace.ParentBased(base)
	}
}
