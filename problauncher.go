package problauncher

import (
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	// 创建 Jaeger Exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://tracing-analysis-dc-bj.aliyuncs.com/adapt_igkxc7d8zo@c3f99692f4e27fe_igkxc7d8zo@53df7ad2afe8301/api/traces")))
	if err != nil {
		panic(err)
		return nil, err
	}

	sdktrace.WithSampler(sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(1.0),
	))

	// 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("echo-server"),
		)),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

var RequestDuration metric.Float64Histogram
var RequestCount metric.Int64Counter

func initMeter() error {
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
	_, err := initTracer()
	if err != nil {
		fmt.Println(err)
	}
	err = initMeter()
	if err != nil {
		fmt.Println(err)
	}
}
