package probesdk

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/trace"
	"strconv"
	"sync"
)

const GRTTraceContextPrefix = "GRTTraceContextKey/"
const GRTTraceContextLen = GRTTraceContextPrefix + "Idx"

const TraceIDSize = 16
const SpanIDSize = 8
const TraceFlagsSize = 1
const RemoteFlagSize = 1 // use one byte to store boolean

const TraceIDStart = 0
const SpanIDStart = TraceIDStart + TraceIDSize
const TraceFlagsStart = SpanIDStart + SpanIDSize
const RemoteFlagStart = TraceFlagsStart + TraceFlagsSize
const TraceStatesStart = RemoteFlagStart + RemoteFlagSize

// 创建内存池
var bytePool = &sync.Pool{
	New: func() interface{} {
		// 每次新建一个长度为26的字节切片
		return make([]byte, 26)
	},
}

func DecodeTraceContext(data string) trace.SpanContext {
	traceState, err := trace.ParseTraceState(data[TraceStatesStart:])
	isRemote := false
	if data[RemoteFlagStart] != 0 {
		isRemote = true
	}
	if err != nil {
		panic(err)
	}
	config := trace.SpanContextConfig{
		TraceState: traceState,
		Remote:     isRemote,
	}
	copy(config.TraceID[:], data[TraceIDStart:SpanIDStart])
	copy(config.SpanID[:], data[SpanIDStart:TraceFlagsStart])
	config.TraceFlags = trace.TraceFlags(data[TraceFlagsStart])
	return trace.NewSpanContext(config)

}

func EncodeTraceContext(ctx trace.SpanContext) string {
	traceStateStr := ctx.TraceState().String()
	bytes := make([]byte, TraceStatesStart+len(traceStateStr))
	traceID := ctx.TraceID()
	spanID := ctx.SpanID()
	copy(bytes[TraceIDStart:SpanIDStart], traceID[:])
	copy(bytes[SpanIDStart:TraceFlagsStart], spanID[:])
	bytes[TraceFlagsStart] = byte(ctx.TraceFlags())
	if ctx.IsRemote() {
		bytes[RemoteFlagStart] = 1
	}
	copy(bytes[TraceStatesStart:], traceStateStr)
	return string(bytes)
}

func getTargetKey(idx int) string {
	return strconv.Itoa(idx)
}

func PopTraceContext() bool {
	data := GetProfLabel()
	sizeStr, _ := data[GRTTraceContextLen]
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return false
	}
	if size == 0 {
		return false
	}
	newSize := size - 1
	removeKey := getTargetKey(newSize)
	newSizeStr := strconv.Itoa(newSize)
	delete(data, removeKey)
	data[GRTTraceContextLen] = newSizeStr
	return true
}

func OnSpanStart(span trace.Span) {
	data := GetProfLabel()
	sizeStr, _ := data[GRTTraceContextLen]
	size, _ := strconv.Atoi(sizeStr)
	addKey := getTargetKey(size)
	newSize := size + 1
	data[GRTTraceContextLen] = strconv.Itoa(newSize)
	data[addKey] = EncodeTraceContext(span.SpanContext())
}

func OnSpanEnd(span trace.Span) bool {
	return PopTraceContext()
}

func RetrieveSpanContext(ctx context.Context) (context.Context, error) {
	data := GetProfLabel()
	sizeStr, _ := data[GRTTraceContextLen]
	size, _ := strconv.Atoi(sizeStr)
	if size == 0 {
		return ctx, fmt.Errorf("no span at top, but RetrieveSpanContext was called")
	}
	targetKey := getTargetKey(size - 1)
	tcStr, _ := data[targetKey]
	return trace.ContextWithSpanContext(ctx, DecodeTraceContext(tcStr)), nil
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
