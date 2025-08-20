package agent

import (
	"context"
	"errors"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/log/global"
	olog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"

	"github.com/complytime/complybeacon/proofwatch/evidence"
)

const name = "go.opentelemetry.io/otel/example/agent"

var (
	meter       = otel.Meter(name)
	serviceName = semconv.ServiceNameKey.String("proofwatch")
)

// otelSDKSetup completes setup of the Otel SDK with providers.
func otelSDKSetup(ctx context.Context, conn *grpc.ClientConn) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	shutDown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			serviceName,
		),
	)
	if err != nil {
		return nil, err
	}

	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(3*time.Second))), sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	logExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	logProcessor := olog.NewSimpleProcessor(logExporter)
	logProvider := olog.NewLoggerProvider(olog.WithProcessor(logProcessor), olog.WithResource(res))

	// Register the provider as the global logger provider.
	global.SetLoggerProvider(logProvider)

	shutdownFuncs = append(shutdownFuncs, logProvider.Shutdown, meterProvider.Shutdown)

	return shutDown, nil
}

func metricsConfigure() *evidence.EvidenceObserver {
	observer, err := evidence.NewEvidenceObserver(meter)
	if err != nil {
		log.Fatalf("failed to register callback: %v", err)
	}
	return observer
}
