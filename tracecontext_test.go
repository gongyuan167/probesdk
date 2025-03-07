package problauncher

import (
	"context"
	"go.opentelemetry.io/otel"
	"testing"
)

func TestSetTraceContext(t *testing.T) {
	ctx := context.Background()
	tracer := otel.GetTracerProvider().Tracer(
		"echo-server",
	)
	ctx1, span1 := tracer.Start(ctx, "level1")
	OnSpanStart(span1)
	ctx2, span2 := tracer.Start(ctx1, "level2")
	OnSpanStart(span2)
	ctx3, span3 := tracer.Start(ctx2, "level3")
	OnSpanStart(span3)
	_, span4 := tracer.Start(ctx3, "level4")
	OnSpanStart(span4)

	latestSpanContext, err := GetSpanContext()
	if err != nil {
		t.Errorf("%v\n", err)
	}

	if latestSpanContext.SpanID() != span4.SpanContext().SpanID() {
		t.Errorf("span id not equal, got %s vs %s", latestSpanContext.SpanID(), span4.SpanContext().SpanID())
	}

	OnSpanEnd(span4)

	latestSpanContext, err = GetSpanContext()
	if err != nil {
		t.Errorf("%v\n", err)
	}

	if latestSpanContext.SpanID() != span3.SpanContext().SpanID() {
		t.Errorf("span id not equal, got %s vs %s", latestSpanContext.SpanID(), span3.SpanContext().SpanID())
	}

	OnSpanEnd(span3)
	latestSpanContext, err = GetSpanContext()
	if err != nil {
		t.Errorf("%v\n", err)
	}

	if latestSpanContext.SpanID() != span2.SpanContext().SpanID() {
		t.Errorf("span id not equal, got %s vs %s", latestSpanContext.SpanID(), span2.SpanContext().SpanID())
	}

	OnSpanEnd(span2)
	latestSpanContext, err = GetSpanContext()
	if err != nil {
		t.Errorf("%v\n", err)
	}

	if latestSpanContext.SpanID() != span1.SpanContext().SpanID() {
		t.Errorf("span id not equal, got %s vs %s", latestSpanContext.SpanID(), span1.SpanContext().SpanID())
	}

	OnSpanEnd(span1)
	latestSpanContext, err = GetSpanContext()
	if err == nil {
		t.Errorf("expect empty span%v\n", err)
	}

}
