package problauncher

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"runtime/pprof"
	"strconv"
	"sync"
)

var DefaultGlobalTraceContext *GlobalTraceContext

const GRTTraceContextKey = "github.com/gongyuan167/problauncher.GRTTraceContextKey"
const DefaultMaxGlobalTraceContext = 100000

type GlobalTraceContext struct {
	data FixedSizeMap[int64, context.Context]
	node *snowflake.Node
	mu   sync.Mutex
}

func InitTraceContext() error {
	var err error
	DefaultGlobalTraceContext, err = NewGlobalTraceContext(DefaultMaxGlobalTraceContext)
	return err
}

func NewGlobalTraceContext(cap int) (*GlobalTraceContext, error) {
	tc := &GlobalTraceContext{
		data: *NewFixedSizeMap[int64, context.Context](cap),
		node: nil,
		mu:   sync.Mutex{},
	}
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, err
	}
	tc.node = node
	return tc, nil
}

func (g *GlobalTraceContext) GenerateUID() int64 {
	id := g.node.Generate()
	return int64(id)
}

func (g *GlobalTraceContext) Store(ctx context.Context) int64 {
	id := g.GenerateUID()
	g.mu.Lock()
	defer g.mu.Unlock()
	g.data.Set(id, ctx)
	return id
}

func (g *GlobalTraceContext) Get(id int64) context.Context {
	g.mu.Lock()
	defer g.mu.Unlock()
	result, ok := g.data.Get(id)
	if !ok {
		return nil
	}
	return result
}

func (g *GlobalTraceContext) Remove(id int64) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.data.Remove(id)
}

func (g *GlobalTraceContext) StoreGLSContext(ctx context.Context) {
	uid := g.Store(ctx)
	lbs := pprof.Labels(GRTTraceContextKey, strconv.FormatInt(uid, 10))
	// todo: it will make the original labels get lost
	gCTX := pprof.WithLabels(context.Background(), lbs)
	pprof.SetGoroutineLabels(gCTX)
}

func (g *GlobalTraceContext) GetGLSContext() context.Context {
	// todo: do we have data race condition here?
	data := GetProfLabel()
	uidStr, ok := data[GRTTraceContextKey]
	if !ok {
		return nil
	}
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid GRTTraceContextKey %s\n", uidStr))
	}
	return g.Get(uid)
}

func (g *GlobalTraceContext) RemoveGLSContext() bool {
	data := GetProfLabel()
	uidStr, ok := data[GRTTraceContextKey]
	if !ok {
		return false
	}
	uid, err := strconv.ParseInt(uidStr, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid GRTTraceContextKey %s\n", uidStr))
	}
	return g.Remove(uid)
}
