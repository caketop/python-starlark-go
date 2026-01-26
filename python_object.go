package main

/*
#include "starlark.h"

extern PyObject *StarlarkError;
extern PyObject *SyntaxError;
extern PyObject *EvalError;
extern PyObject *ResolveError;
*/
import "C"

import (
	"fmt"
	"runtime/cgo"
	"sync"
	"unsafe"

	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
)

type StarlarkState struct {
	Globals     starlark.StringDict
	Mutex       sync.RWMutex
	Print       *C.PyObject
	threadState *C.PyThreadState
	// Most Python values are copied into a new starlark.Value, including
	// lists, dicts, sets, etc. But some values, namely functions, keep a
	// reference to the original function, so we need to INCREF the function
	// and DECREF when Starlark is deallocated.
	//
	// Currently, we only DECREF everything on deallocation, so there's a
	// memory leak if someone keeps replacing the same global with a different
	// function, but that should be rare and would make the implementation more
	// difficult.
	childRefs   []*C.PyObject
}

//export ConfigureStarlark
func ConfigureStarlark(allowSet C.int, allowGlobalReassign C.int, allowRecursion C.int) {
	// Ignore input values other than 0 or 1 and leave current value in place
	switch allowSet {
	case 0:
		resolve.AllowSet = false
	case 1:
		resolve.AllowSet = true
	}

	switch allowGlobalReassign {
	case 0:
		resolve.AllowGlobalReassign = false
	case 1:
		resolve.AllowGlobalReassign = true
	}

	switch allowRecursion {
	case 0:
		resolve.AllowRecursion = false
	case 1:
		resolve.AllowRecursion = true
	}
}

func rlockSelf(self *C.Starlark) *StarlarkState {
	state := cgo.Handle(self.handle).Value().(*StarlarkState)
	state.Mutex.RLock()
	return state
}

func lockSelf(self *C.Starlark) *StarlarkState {
	state := cgo.Handle(self.handle).Value().(*StarlarkState)
	state.Mutex.Lock()
	return state
}

func (state *StarlarkState) DetachGIL() {
	state.threadState = C.PyEval_SaveThread()
}

func (state *StarlarkState) ReattachGIL() {
	if state.threadState == nil {
		return
	}

	C.PyEval_RestoreThread(state.threadState)
	state.threadState = nil
}

//export Starlark_new
func Starlark_new(pytype *C.PyTypeObject, args *C.PyObject, kwargs *C.PyObject) *C.Starlark {
	self := C.starlarkAlloc(pytype)
	if self == nil {
		return nil
	}

	state := &StarlarkState{
		Globals: starlark.StringDict{},
		Mutex: sync.RWMutex{},
		Print: nil,
		threadState: nil,
	}
	self.handle = C.uintptr_t(cgo.NewHandle(state))

	return self
}

//export Starlark_init
func Starlark_init(self *C.Starlark, args *C.PyObject, kwargs *C.PyObject) C.int {
	var globals *C.PyObject = nil
	var print *C.PyObject = nil

	if C.parseInitArgs(args, kwargs, &globals, &print) == 0 {
		return -1
	}

	if print != nil {
		if Starlark_set_print(self, print, nil) != 0 {
			return -1
		}
	}

	if globals != nil {
		if C.PyMapping_Check(globals) != 1 {
			errmsg := C.CString(fmt.Sprintf("Can't initialize globals from %s", C.GoString(globals.ob_type.tp_name)))
			defer C.free(unsafe.Pointer(errmsg))
			C.PyErr_SetString(C.PyExc_TypeError, errmsg)
			return -1
		}

		retval := Starlark_set_globals(self, args, globals)
		if retval == nil {
			return -1
		}
	}

	return 0
}

//export Starlark_dealloc
func Starlark_dealloc(self *C.Starlark) {
	handle := cgo.Handle(self.handle)
	state := handle.Value().(*StarlarkState)

	handle.Delete()

	state.Mutex.Lock()
	defer state.Mutex.Unlock()

	for _, obj := range state.childRefs {
		C.Py_DecRef(obj)
	}

	if state.Print != nil {
		C.Py_DecRef(state.Print)
	}

	C.starlarkFree(self)
}

func main() {}
