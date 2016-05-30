package plugger

// #include <stdlib.h>
// #include <dlfcn.h>
// #cgo LDFLAGS: -ldl   
import "C"

import (
	"fmt"
	"unsafe"
)

type Flags int
const (
	Lazy Flags = C.RTLD_LAZY
	Now Flags = C.RTLD_NOW
	Global Flags = C.RTLD_GLOBAL
	Local Flags = C.RTLD_LOCAL
	NoLoad Flags = C.RTLD_NOLOAD
	NoDelete Flags = C.RTLD_NODELETE
)

type dlHandle struct {
	dlPtr unsafe.Pointer
}

func dlOpen(fname string, flags Flags) (*dlHandle, error) {
	csfname := C.CString(fname)
	defer C.free(unsafe.Pointer(csfname))

	h := C.dlopen(csfname, C.int(flags))
	if h == nil {
		cErr := C.dlerror()
		return nil, fmt.Errorf("dlOpen: %s", C.GoString(cErr))
	}
	return &dlHandle{dlPtr: h}, nil
}

func (dh *dlHandle) dlClose() error {
	o := C.dlclose(dh.dlPtr)
	if o != C.int(0) {
		cErr := C.dlerror()
		return fmt.Errorf("dlClose: %s", C.GoString(cErr))
	}
	return nil
}

func (dh *dlHandle) dlSymbol(symbol string) (unsafe.Pointer, error) {
	cssym := C.CString(symbol)
	defer C.free(unsafe.Pointer(cssym))
	symPtr := C.dlsym(dh.dlPtr, cssym)
	if symPtr == nil {
		cErr := C.dlerror()
		return nil, fmt.Errorf("dlSymbol: %s", C.GoString(cErr))
	}
	return symPtr, nil
}
