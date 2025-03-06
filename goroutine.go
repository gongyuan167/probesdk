package problauncher

import (
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
		return map[string]string{}
	}
	return *result
}
