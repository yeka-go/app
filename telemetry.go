package app

import (
	"context"
	"errors"
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
	ServiceName string `mapstructure:"service_name"`

	Tracer observabilityConfig `mapstructure:"tracer"`
	Metric observabilityConfig `mapstructure:"metric"`
	Logger observabilityConfig `mapstructure:"logger"`

	Exporters map[string]exporterConfig `mapstructure:"exporters"`
}

type observabilityConfig struct {
	Exporter string `mapstructure:"exporter"`
}

type exporterConfig struct {
	OtelGRPCAddr string            `mapstructure:"otlp_grpc_host"`
	Headers      map[string]string `mapstructure:"headers"`
	Insecure     bool              `mapstructure:"insecure"`
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
	if cfg.Metric.Exporter == "" {
		return nil, nil
	}
	exp, ok := cfg.Exporters[cfg.Metric.Exporter]
	if !ok {
		return nil, errors.New("telemetry.metric.exporter not found: " + cfg.Metric.Exporter)
	}

	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithCompressor("gzip"),
	}

	if exp.OtelGRPCAddr != "" {
		opts = append(opts, otlpmetricgrpc.WithEndpoint(exp.OtelGRPCAddr))
	}
	if exp.Insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	if len(exp.Headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(exp.Headers))
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

	if cfg.Logger.Exporter == "" {
		return nil, nil
	}
	exp, ok := cfg.Exporters[cfg.Logger.Exporter]
	if !ok {
		return nil, errors.New("telemetry.logger.exporter not found: " + cfg.Logger.Exporter)
	}

	opts := []otlploggrpc.Option{
		otlploggrpc.WithCompressor("gzip"),
	}

	if exp.OtelGRPCAddr != "" {
		opts = append(opts, otlploggrpc.WithEndpoint(exp.OtelGRPCAddr))
	}
	if exp.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	if len(exp.Headers) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(exp.Headers))
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

	handler = append(handler, otelslog.NewHandler("github.com/yeka-go/app"))
	return provider.Shutdown, nil
}

func initTracer(cfg telemetryConfig, commonResource *resource.Resource) (ShutdownFunc, error) {
	if cfg.Tracer.Exporter == "" {
		return nil, nil
	}
	exp, ok := cfg.Exporters[cfg.Tracer.Exporter]
	if !ok {
		return nil, errors.New("telemetry.tracer.exporter not found: " + cfg.Tracer.Exporter)
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithCompressor("gzip"),
	}

	if exp.OtelGRPCAddr != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(exp.OtelGRPCAddr))
	}
	if exp.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if len(exp.Headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(exp.Headers))
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
