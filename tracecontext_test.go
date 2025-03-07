package problauncher

import (
	"context"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"testing"
)

func GetSpanContext() (trace.SpanContext, error) {
	ctx := context.Background()
	ctx, err := RetrieveSpanContext(ctx)
	return trace.SpanContextFromContext(ctx), err
}

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

// 初始化TracerProvider
func initTracer() *sdktrace.TracerProvider {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	return tp
}

// 独立封装的Span启动函数
func startSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	nctx, _ := RetrieveSpanContext(context.Background())
	newCtx, span := otel.Tracer("benchmark").Start(nctx, name)
	OnSpanStart(span)
	return newCtx, span
}

// 新增封装的Span结束函数
func endSpan(span trace.Span) {
	OnSpanEnd(span)
	span.End()
}

func BenchmarkNestedSpans(b *testing.B) {
	tp := initTracer()
	defer tp.Shutdown(context.Background())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建根Span
		ctx, root := startSpan(context.Background(), "root")

		// 第一层子Span
		childCtx1, child1 := startSpan(ctx, "child1")
		_, grandChild1 := startSpan(childCtx1, "grandchild1")
		endSpan(grandChild1) // 使用封装函数结束
		endSpan(child1)

		// 第二层子Span
		childCtx2, child2 := startSpan(ctx, "child2")

		// 嵌套两层
		grandChildCtx2, grandChild2 := startSpan(childCtx2, "grandchild2")
		_, greatGrandChild := startSpan(grandChildCtx2, "great-grandchild")
		endSpan(greatGrandChild)
		endSpan(grandChild2)

		// 创建同级Span
		_, grandChild3 := startSpan(childCtx2, "grandchild3")
		endSpan(grandChild3)

		endSpan(child2)

		// 复用上下文创建新Span
		revivedSpanCtx, revivedSpan := startSpan(grandChildCtx2, "revived")
		_, deepChild := startSpan(revivedSpanCtx, "deepChild")
		endSpan(deepChild)
		endSpan(revivedSpan)

		endSpan(root) // 结束根Span
	}
}

// 独立封装的Span启动函数
func directStartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer("benchmark").Start(ctx, name)
}

// 新增封装的Span结束函数
func directEndSpan(span trace.Span) {
	span.End()
}

func BenchmarkDirectNestedSpans(b *testing.B) {
	tp := initTracer()
	defer tp.Shutdown(context.Background())

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 创建根Span
		ctx, root := directStartSpan(context.Background(), "root")

		// 第一层子Span
		childCtx1, child1 := directStartSpan(ctx, "child1")
		_, grandChild1 := directStartSpan(childCtx1, "grandchild1")
		directEndSpan(grandChild1) // 使用封装函数结束
		directEndSpan(child1)

		// 第二层子Span
		childCtx2, child2 := directStartSpan(ctx, "child2")

		// 嵌套两层
		grandChildCtx2, grandChild2 := directStartSpan(childCtx2, "grandchild2")
		_, greatGrandChild := directStartSpan(grandChildCtx2, "great-grandchild")
		directEndSpan(greatGrandChild)
		directEndSpan(grandChild2)

		// 创建同级Span
		_, grandChild3 := directStartSpan(childCtx2, "grandchild3")
		directEndSpan(grandChild3)

		directEndSpan(child2)

		// 复用上下文创建新Span
		revivedSpanCtx, revivedSpan := directStartSpan(grandChildCtx2, "revived")
		_, deepChild := directStartSpan(revivedSpanCtx, "deepChild")
		directEndSpan(deepChild)
		directEndSpan(revivedSpan)

		directEndSpan(root) // 结束根Span
	}
}
