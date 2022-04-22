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
	"math/rand"
	"sync"
	"unsafe"

	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
)

type StarlarkState struct {
	Globals starlark.StringDict
	Mutex   sync.RWMutex
}

var (
	STATE       = map[uint64]*StarlarkState{}
	STATE_MUTEX = sync.Mutex{}
)

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
	stateId := uint64(self.state_id)
	state, ok := STATE[stateId]

	if !ok {
		raiseRuntimeError("Internal error: rlockSelf: unknown state_id")
		return nil
	}

	state.Mutex.RLock()
	return state
}

func lockSelf(self *C.Starlark) *StarlarkState {
	stateId := uint64(self.state_id)
	state, ok := STATE[stateId]

	if !ok {
		raiseRuntimeError("Internal error: lockSelf: unknown state_id")
		return nil
	}

	state.Mutex.Lock()
	return state
}

//export Starlark_new
func Starlark_new(pytype *C.PyTypeObject, args *C.PyObject, kwargs *C.PyObject) *C.Starlark {
	self := C.starlarkAlloc(pytype)
	if self == nil {
		return nil
	}

	var stateId uint64

	STATE_MUTEX.Lock()
	defer STATE_MUTEX.Unlock()

	for {
		stateId = rand.Uint64()
		_, ok := STATE[stateId]
		if !ok {
			break
		}
	}

	self.state_id = C.ulong(stateId)
	STATE[stateId] = &StarlarkState{Globals: starlark.StringDict{}, Mutex: sync.RWMutex{}}
	return self
}

//export Starlark_init
func Starlark_init(self *C.Starlark, args *C.PyObject, kwargs *C.PyObject) C.int {
	var globals *C.PyObject = nil
	if C.parseInitArgs(args, kwargs, &globals) == 0 {
		return -1
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
	STATE_MUTEX.Lock()
	defer STATE_MUTEX.Unlock()

	stateId := uint64(self.state_id)
	state := STATE[stateId]

	state.Mutex.Lock()
	defer state.Mutex.Unlock()

	delete(STATE, stateId)
	C.starlarkFree(self)
}

func main() {}
