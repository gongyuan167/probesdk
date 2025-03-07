package problauncher

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"runtime/pprof"
)

// var DefaultGlobalTraceContext *GlobalTraceContext

const GRTTraceContextKey = "github.com/gongyuan167/problauncher.GRTTraceContextKey"

// const DefaultMaxGlobalTraceContext = 100000

type LinkedTraceContext struct {
	TraceContext trace.SpanContextConfig
	Parent       *LinkedTraceContext
}

// SavedTraceContext JSON中存储的数据格式
type SavedTraceContext struct {
	TraceID    string
	SpanID     string
	TraceFlags string
	TraceState string
	Remote     bool
}

// SavedLinkedTraceContext 便于解析的数据格式
type SavedLinkedTraceContext struct {
	TraceContext SavedTraceContext
	Parent       *SavedLinkedTraceContext
}

func getGRTLinkedTraceContext() *SavedLinkedTraceContext {
	data := GetProfLabel()
	jsonStr, ok := data[GRTTraceContextKey]
	if !ok || len(jsonStr) == 0 {
		return nil
	}
	result := &SavedLinkedTraceContext{}
	err := json.Unmarshal([]byte(jsonStr), result)
	if err != nil {
		panic(err)
	}
	return result
}

func setGRTLabels(str string) {
	lbs := pprof.Labels(GRTTraceContextKey, str)
	// todo: it will make the original labels get lost
	gCTX := pprof.WithLabels(context.Background(), lbs)
	pprof.SetGoroutineLabels(gCTX)
}

func (l *SavedLinkedTraceContext) setGRTLinkedTraceContext() {
	jsonStr, err := json.Marshal(l)
	if err != nil {
		panic(err)
	}
	setGRTLabels(string(jsonStr))
}

func OnSpanStart(span trace.Span) {
	parent := getGRTLinkedTraceContext()
	stx := span.SpanContext()
	savedContext := SavedTraceContext{
		TraceID:    stx.TraceID().String(),
		SpanID:     stx.SpanID().String(),
		TraceFlags: stx.TraceFlags().String(),
		TraceState: stx.TraceState().String(),
		Remote:     stx.IsRemote(),
	}

	ctx := SavedLinkedTraceContext{
		TraceContext: savedContext,
		Parent:       parent,
	}

	ctx.setGRTLinkedTraceContext()
}

func OnSpanEnd(span trace.Span) {
	top := getGRTLinkedTraceContext()
	stx := span.SpanContext()
	if top == nil || top.TraceContext.SpanID != stx.SpanID().String() {
		fmt.Errorf("wrong span configuration detected!\b")
		setGRTLabels("")
		return
	}
	if top.Parent == nil {
		setGRTLabels("")
		return
	}
	top.Parent.setGRTLinkedTraceContext()
}

func GetSpanContext() (trace.SpanContext, error) {
	result := getGRTLinkedTraceContext()
	if result == nil {
		return trace.SpanContext{}, fmt.Errorf("empty span context")
	}
	data := result.TraceContext
	// 转换TraceID
	traceID, err := trace.TraceIDFromHex(data.TraceID)
	if err != nil {
		panic(fmt.Errorf("invalid TraceID: %v", err))
	}

	// 转换SpanID
	spanID, err := trace.SpanIDFromHex(data.SpanID)
	if err != nil {
		panic(fmt.Errorf("invalid SpanID: %v", err))
	}

	// 转换TraceFlags
	traceFlagsBytes, err := hex.DecodeString(data.TraceFlags)
	if err != nil || len(traceFlagsBytes) != 1 {
		panic(fmt.Errorf("invalid TraceFlags: %v", err))
	}
	traceFlags := trace.TraceFlags(traceFlagsBytes[0])

	// 转换TraceState（若为空则返回空状态）
	traceState, err := trace.ParseTraceState(data.TraceState)
	if err != nil {
		panic(fmt.Errorf("invalid TraceState: %v", err))
	}

	// 创建SpanContext
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: traceFlags,
		TraceState: traceState,
		Remote:     data.Remote,
	})

	// 验证SpanContext是否有效
	if !spanContext.IsValid() {
		panic("resulting SpanContext is invalid")
	}
	return spanContext, nil
}

func RetrieveSpanContext(ctx context.Context) (context.Context, error) {
	spanCtx, err := GetSpanContext()
	if err != nil {
		return ctx, err
	}
	return trace.ContextWithSpanContext(ctx, spanCtx), nil
}

//type GlobalTraceContext struct {
//	data FixedSizeMap[int64, context.Context]
//	node *snowflake.Node
//	mu   sync.Mutex
//}
//
//func InitTraceContext() error {
//	var err error
//	DefaultGlobalTraceContext, err = NewGlobalTraceContext(DefaultMaxGlobalTraceContext)
//	return err
//}
//
//func NewGlobalTraceContext(cap int) (*GlobalTraceContext, error) {
//	tc := &GlobalTraceContext{
//		data: *NewFixedSizeMap[int64, context.Context](cap),
//		node: nil,
//		mu:   sync.Mutex{},
//	}
//	node, err := snowflake.NewNode(1)
//	if err != nil {
//		return nil, err
//	}
//	tc.node = node
//	return tc, nil
//}
//
//func (g *GlobalTraceContext) GenerateUID() int64 {
//	id := g.node.Generate()
//	return int64(id)
//}
//
//func (g *GlobalTraceContext) Store(ctx context.Context) int64 {
//	id := g.GenerateUID()
//	g.mu.Lock()
//	defer g.mu.Unlock()
//	g.data.Set(id, ctx)
//	return id
//}
//
//func (g *GlobalTraceContext) Get(id int64) context.Context {
//	g.mu.Lock()
//	defer g.mu.Unlock()
//	result, ok := g.data.Get(id)
//	if !ok {
//		return nil
//	}
//	return result
//}
//
//func (g *GlobalTraceContext) Remove(id int64) bool {
//	g.mu.Lock()
//	defer g.mu.Unlock()
//	return g.data.Remove(id)
//}
//
//func (g *GlobalTraceContext) StoreGLSContext(ctx context.Context) {
//	uid := g.Store(ctx)
//	lbs := pprof.Labels(GRTTraceContextKey, strconv.FormatInt(uid, 10))
//	// todo: it will make the original labels get lost
//	gCTX := pprof.WithLabels(context.Background(), lbs)
//	pprof.SetGoroutineLabels(gCTX)
//}
//
//func (g *GlobalTraceContext) GetGLSContext() context.Context {
//	// todo: do we have data race condition here?
//	data := GetProfLabel()
//	uidStr, ok := data[GRTTraceContextKey]
//	if !ok {
//		return nil
//	}
//	uid, err := strconv.ParseInt(uidStr, 10, 64)
//	if err != nil {
//		panic(fmt.Sprintf("invalid GRTTraceContextKey %s\n", uidStr))
//	}
//	return g.Get(uid)
//}
//
//func (g *GlobalTraceContext) RemoveGLSContext() bool {
//	data := GetProfLabel()
//	uidStr, ok := data[GRTTraceContextKey]
//	if !ok {
//		return false
//	}
//	uid, err := strconv.ParseInt(uidStr, 10, 64)
//	if err != nil {
//		panic(fmt.Sprintf("invalid GRTTraceContextKey %s\n", uidStr))
//	}
//	return g.Remove(uid)
//}
