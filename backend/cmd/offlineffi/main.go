// Build with -buildmode=c-shared. Experimental P01 ABI; not a network API.
package main

/*
#include <stdlib.h>
#include <stdint.h>
*/
import "C"

import (
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/offlineprobe"
	"unsafe"
)

// CGMSProbeAlloc and CGMSProbeFree use one allocator across the ABI boundary.
// Caller owns every allocation and must free it exactly once after the call.
//
//export CGMSProbeAlloc
func CGMSProbeAlloc(size C.int32_t) unsafe.Pointer {
	if size < 1 || size > canonical.MaxBytes {
		return nil
	}
	return C.malloc(C.size_t(size))
}

//export CGMSProbeFree
func CGMSProbeFree(ptr unsafe.Pointer) { C.free(ptr) }

// CGMSProbe copies a UTF-8 request into Go and writes to caller-owned memory.
// Return: positive output byte length, -1 invalid buffers, -2 insufficient output
// capacity, -3 recovered internal panic. Buffers must be valid and non-overlapping
// for their declared sizes. No Go pointer or retained state crosses the boundary.
//
//export CGMSProbe
func CGMSProbe(input unsafe.Pointer, length C.int32_t, output unsafe.Pointer, capacity C.int32_t) (result C.int32_t) {
	result = -3
	defer func() {
		if recover() != nil {
			result = -3
		}
	}()
	if input == nil || output == nil || length < 1 || length > canonical.MaxBytes || capacity < 1 || capacity > canonical.MaxBytes {
		return -1
	}
	b := offlineprobe.Execute(C.GoBytes(input, C.int(length)))
	if len(b) > int(capacity) {
		return -2
	}
	copy(unsafe.Slice((*byte)(output), int(capacity)), b)
	return C.int32_t(len(b))
}
func main() {}
