package problauncher

import (
	"context"
	"runtime/pprof"
	_ "runtime/pprof"
	"unsafe"
)

// @see: runtime/pprof/label.go
type labelMap map[string]string

//go:linkname Runtime_getProfLabel runtime/pprof.runtime_getProfLabel
func Runtime_getProfLabel() unsafe.Pointer

func GetProfLabel() map[string]string {
	ptr := Runtime_getProfLabel()
	result := (*labelMap)(ptr)
	if result == nil {
		ctx := pprof.WithLabels(context.Background(), pprof.Labels())
		pprof.SetGoroutineLabels(ctx)
		return GetProfLabel()
	}
	return *result
}
