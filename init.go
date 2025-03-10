package probesdk

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	metric2 "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const HTTP_ENDPOINT = "tracing-analysis-dc-bj.aliyuncs.com"
const HTTP_TRACE_URL_PATH = "adapt_igkxc7d8zo@c3f99692f4e27fe_igkxc7d8zo@53df7ad2afe8301/api/otlp/traces"

const HTTP_METRIC_ENDPOINT = "cn-beijing.arms.aliyuncs.com"
const HTTP_METRICS_URL_PATH = "opentelemetry/58f1a59e132c474b139cf8e4366552/1874856833619396/i8anrmcvv6/cn-beijing/api/v1/metrics"

// 设置应用资源
func newResource(ctx context.Context) *resource.Resource {
	hostName, _ := os.Hostname()
	programName := filepath.Base(os.Args[0])
	programName = strings.TrimSuffix(programName, filepath.Ext(programName))

	r, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.HostNameKey.String(hostName), // 主机名
			semconv.ServiceNameKey.String(programName),
		),
	)

	if err != nil {
		log.Fatalf("%s: %v", "Failed to create OpenTelemetry resource", err)
	}
	return r
}

func newHTTPExporterAndSpanProcessor(ctx context.Context) (*otlptrace.Exporter, sdktrace.SpanProcessor) {

	traceExporter, err := otlptrace.New(ctx, otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint(HTTP_ENDPOINT),
		otlptracehttp.WithURLPath(HTTP_TRACE_URL_PATH),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithCompression(1)))

	if err != nil {
		log.Fatalf("%s: %v", "Failed to create the OpenTelemetry trace exporter", err)
	}

	batchSpanProcessor := sdktrace.NewBatchSpanProcessor(traceExporter)

	return traceExporter, batchSpanProcessor
}

// InitOpenTelemetryTrace  OpenTelemetry 初始化方法
func InitOpenTelemetryTrace(ctx context.Context, otelResource *resource.Resource) func() {

	var traceExporter *otlptrace.Exporter
	var batchSpanProcessor sdktrace.SpanProcessor

	traceExporter, batchSpanProcessor = newHTTPExporterAndSpanProcessor(ctx)

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(otelResource),
		sdktrace.WithSpanProcessor(batchSpanProcessor))

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := traceExporter.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}
}

var RequestDuration metric.Float64Histogram
var RequestCount metric.Int64Counter

func initMeter(ctx context.Context, otelResource *resource.Resource) error {
	exporter, err := otlpmetrichttp.New(
		ctx,
		otlpmetrichttp.WithEndpoint(HTTP_METRIC_ENDPOINT),
		otlpmetrichttp.WithURLPath(HTTP_METRICS_URL_PATH),
		otlpmetrichttp.WithInsecure(),
		//otlpmetrichttp.WithHeaders(map[string]string{
		//	"Authorization": "Bearer YOUR_TOKEN", // 阿里云要求的认证头
		//}),
	)
	if err != nil {
		panic(err)
	}

	// 4. 注册全局 MeterProvider
	// 3. 创建 MeterProvider
	provider := metric2.NewMeterProvider(
		metric2.WithResource(otelResource),
		metric2.WithReader(metric2.NewPeriodicReader(
			exporter, metric2.WithInterval(5*time.Second))))

	otel.SetMeterProvider(provider)
	meter := otel.Meter("")
	RequestDuration, _ = meter.Float64Histogram(
		"http.request.duration",
		metric.WithUnit("ms"),
	)
	RequestCount, _ = meter.Int64Counter(
		"http.request.count",
	)

	return nil
}

func init() {
	fmt.Println(".............Init Probe ...........")
	ctx := context.Background()
	otelResource := newResource(ctx)

	_ = InitOpenTelemetryTrace(ctx, otelResource)

	err := initMeter(ctx, otelResource)
	if err != nil {
		fmt.Println(err)
	}

}
