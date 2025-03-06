package problauncher

import (
	"context"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"runtime/pprof"
	"testing"
)

func TestGetProfLabel(t *testing.T) {
	ctx := context.Background()
	lbs := pprof.Labels("key1", "val1", "key2", "val2")
	pprof.Do(ctx, lbs, func(ctx context.Context) {
		if len(GetProfLabel()) != 2 {
			t.Fail()
		}
	})
}

func TestTraceContext(t *testing.T) {
	ctx := context.Background()
	tracer := otel.GetTracerProvider().Tracer(
		"echo-server",
		trace.WithInstrumentationVersion("v4.0.0"),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	ctx, span := tracer.Start(
		ctx,
		"level1",
		trace.WithSpanKind(trace.SpanKindServer),
	)

	spanRecovered := trace.SpanFromContext(ctx)
	if spanRecovered == nil || spanRecovered.SpanContext().SpanID() != span.SpanContext().SpanID() {
		t.Fail()
	}
	defer span.End()

}
