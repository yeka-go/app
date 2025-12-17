package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	slogmulti "github.com/samber/slog-multi"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log/global"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type telemetryConfig struct {
	ServiceName  string            `mapstructure:"service_name"`
	OtelGRPCAddr string            `mapstructure:"otlp_grpc_host"`
	Headers      map[string]string `mapstructure:"headers"`
	Insecure     bool              `mapstructure:"insecure"`

	Tracer bool `mapstructure:"tracer"`
	Metric bool `mapstructure:"metric"`
	Logger bool `mapstructure:"logger"`
}

func initTelemetry(config *viper.Viper) error {
	cfg := telemetryConfig{}
	err := config.UnmarshalKey("telemetry", &cfg)
	if err != nil {
		return fmt.Errorf("config.Unmarshal: %w", err)
	}

	commonResource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
	)

	for _, initFn := range []func(telemetryConfig, *resource.Resource) (ShutdownFunc, error){initTracer, initLogger, initMeter} {
		fn, err := initFn(cfg, commonResource)
		if err != nil {
			return err
		}
		if fn != nil {
			OnShutdown(fn)
		}
	}
	return nil
}

func initMeter(cfg telemetryConfig, commonResource *resource.Resource) (ShutdownFunc, error) {
	if !cfg.Metric {
		return nil, nil
	}

	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithCompressor("gzip"),
	}

	if cfg.OtelGRPCAddr != "" {
		opts = append(opts, otlpmetricgrpc.WithEndpoint(cfg.OtelGRPCAddr))
	}
	if cfg.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(cfg.Headers))
	}

	exporter, err := otlpmetricgrpc.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("initMeter: %w", err)
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(
				exporter,
				// Default is 1m. Set to 3s for demonstrative purposes.
				metric.WithInterval(3*time.Second),
			),
		),
		metric.WithResource(commonResource),
	)
	otel.SetMeterProvider(provider)

	// DEFAULT metrics
	// meter := otel.Meter("core/meter")
	// memUsageGauge, _ := meter.Float64ObservableGauge("memory_usage_bytes")
	// cpuUsageGauge, _ := meter.Float64ObservableGauge("cpu_usage_percent")
	// routineGauge, _ := meter.Int64ObservableCounter("num_go_routine")

	// _, err = meter.RegisterCallback(
	// 	func(ctx context.Context, o otelmetric.Observer) error {
	// 		var memStats runtime.MemStats
	// 		runtime.ReadMemStats(&memStats)
	// 		o.ObserveFloat64(memUsageGauge, float64(memStats.Alloc))

	// 		percentages, err := cpu.Percent(0, false)
	// 		if err == nil && len(percentages) > 0 {
	// 			o.ObserveFloat64(cpuUsageGauge, percentages[0])
	// 		}

	// 		o.ObserveInt64(routineGauge, int64(runtime.NumGoroutine()))
	// 		return nil
	// 	},
	// 	memUsageGauge, cpuUsageGauge, routineGauge,
	// )

	otelruntime.Start(otelruntime.WithMinimumReadMemStatsInterval(time.Second))

	return provider.Shutdown, nil
}

func initLogger(cfg telemetryConfig, commonResource *resource.Resource) (ShutdownFunc, error) {
	var exporter logsdk.Exporter
	var err error

	handler := []slog.Handler{
		slog.NewTextHandler(os.Stdout, nil),
	}
	defer func() {
		slog.SetDefault(slog.New(slogmulti.Fanout(handler...)))
	}()

	if !cfg.Logger {
		return func(ctx context.Context) error { return nil }, nil
	}

	opts := []otlploggrpc.Option{
		otlploggrpc.WithCompressor("gzip"),
	}

	if cfg.OtelGRPCAddr != "" {
		opts = append(opts, otlploggrpc.WithEndpoint(cfg.OtelGRPCAddr))
	}
	if cfg.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(cfg.Headers))
	}

	exporter, err = otlploggrpc.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("initLogger: %w", err)
	}

	provider := logsdk.NewLoggerProvider(
		logsdk.WithProcessor(logsdk.NewBatchProcessor(exporter)),
		logsdk.WithResource(commonResource),
	)
	global.SetLoggerProvider(provider)

	handler = append(handler, otelslog.NewHandler("core/logger"))
	return provider.Shutdown, nil
}

func initTracer(cfg telemetryConfig, commonResource *resource.Resource) (ShutdownFunc, error) {
	if !cfg.Tracer {
		return nil, nil
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithCompressor("gzip"),
	}

	if cfg.OtelGRPCAddr != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(cfg.OtelGRPCAddr))
	}
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(cfg.Headers))
	}

	ctx := context.Background()
	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("initTracer: %w", err)
	}

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(commonResource),
	)
	otel.SetTracerProvider(provider)

	otel.SetTextMapPropagator(b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)))

	return provider.Shutdown, nil
}
