package problauncher

import (
	"context"
	"fmt"
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

func TestTraceTree(t *testing.T) {
	ctx := context.Background()
	tracer := otel.GetTracerProvider().Tracer(
		"echo-server",
		trace.WithInstrumentationVersion("v4.0.0"),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
	ctx1, span1 := tracer.Start(ctx, "level1")
	ctx2, span2 := tracer.Start(ctx1, "level2")
	ctx3, span3 := tracer.Start(ctx2, "level3")
	ctx4, span4 := tracer.Start(ctx3, "level4")

	ctx5 := context.WithValue(ctx4, "111", "222")

	fmt.Println(span1)
	fmt.Println(span2)
	fmt.Println(span3)
	fmt.Println(span4)
	fmt.Println(ctx4)
	fmt.Println(ctx5)
}
