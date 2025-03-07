package problauncher

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"go.opentelemetry.io/otel/trace"
	"strconv"
)

// var DefaultGlobalTraceContext *GlobalTraceContext

const GRTTraceContextPrefix = "GRTTraceContextKey/"
const GRTTraceContextLen = GRTTraceContextPrefix + "Idx"

func NewTraceContext(ctx trace.SpanContext) *TraceContextProto {
	traceID := ctx.TraceID()
	spanID := ctx.SpanID()
	traceFlags := byte(ctx.TraceFlags())
	tcp := TraceContextProto{
		TraceId:    traceID[:],
		SpanId:     spanID[:],
		TraceFlags: []byte{traceFlags},
		TraceState: ctx.TraceState().String(),
		Remote:     ctx.IsRemote(),
	}
	return &tcp
}

func DecodeTraceContext(data string) *TraceContextProto {
	result := &TraceContextProto{}
	err := proto.Unmarshal([]byte(data), result)
	if err != nil {
		panic(err)
	}
	return result
}

func EncodeTraceContext(tc *TraceContextProto) string {
	result, err := proto.Marshal(tc)
	if err != nil {
		panic(err)
	}
	return string(result)
}

func getTargetKey(idx int) string {
	return GRTTraceContextPrefix + strconv.Itoa(idx)
}

func TopTraceContext() (*TraceContextProto, bool) {
	data := GetProfLabel()
	sizeStr, _ := data[GRTTraceContextLen]
	size, _ := strconv.Atoi(sizeStr)
	if size == 0 {
		return nil, false
	}
	targetKey := getTargetKey(size - 1)
	tcStr, _ := data[targetKey]
	tc := DecodeTraceContext(tcStr)
	return tc, true
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

func PushTraceContext(tc *TraceContextProto) {
	data := GetProfLabel()
	sizeStr, _ := data[GRTTraceContextLen]
	size, _ := strconv.Atoi(sizeStr)
	addKey := getTargetKey(size)
	newSize := size + 1
	data[GRTTraceContextLen] = strconv.Itoa(newSize)
	data[addKey] = EncodeTraceContext(tc)
}

func OnSpanStart(span trace.Span) {
	tc := NewTraceContext(span.SpanContext())
	PushTraceContext(tc)
}

func OnSpanEnd(span trace.Span) bool {
	return PopTraceContext()
}

func RetrieveSpanContext(ctx context.Context) (context.Context, error) {
	tc, ok := TopTraceContext()
	if !ok {
		return ctx, fmt.Errorf("no tracing context found")
	}
	traceState, err := trace.ParseTraceState(tc.TraceState)
	if err != nil {
		return ctx, err
	}
	traceCfg := trace.SpanContextConfig{
		TraceFlags: trace.TraceFlags(tc.TraceFlags[0]),
		TraceState: traceState,
		Remote:     tc.Remote,
	}
	copy(traceCfg.TraceID[:], tc.TraceId)
	copy(traceCfg.SpanID[:], tc.SpanId)
	return trace.ContextWithSpanContext(ctx, trace.NewSpanContext(traceCfg)), nil
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
