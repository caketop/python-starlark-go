package main

/*
#include "starlark.h"
*/
import "C"

import (
	"sync/atomic"
	"time"
	"unsafe"

	"go.starlark.net/starlark"
)

// Manage the Python global interpreter lock (GIL)
type PythonEnv interface {
	// Detach the GIL and save the thread state
	DetachGIL()

	// Re-attach the GIL with the saved thread state
	ReattachGIL()
}

//export Starlark_eval
func Starlark_eval(self *C.Starlark, args *C.PyObject, kwargs *C.PyObject) *C.PyObject {
	var (
		expr       *C.char
		filename   *C.char     = nil
		convert    C.uint      = 1
		print      *C.PyObject = nil
		timeout    C.double    = 0
		goFilename string      = "<expr>"
	)

	if C.parseEvalArgs(args, kwargs, &expr, &filename, &convert, &print, &timeout) == 0 {
		return nil
	}

	print = pythonPrint(self, print)
	if print == nil {
		return nil
	}

	goExpr := C.GoString(expr)
	if filename != nil {
		goFilename = C.GoString(filename)
	}

	state := rlockSelf(self)
	if state == nil {
		return nil
	}
	defer state.Mutex.RUnlock()

	state.DetachGIL()
	starlarkPrint := func(_ *starlark.Thread, msg string) {
		state.ReattachGIL()
		defer state.DetachGIL()

		callPythonPrint(print, msg)
	}

	thread := &starlark.Thread{Print: starlarkPrint}

	var timedOut atomic.Bool
	if timeout > 0 {
		timer := time.AfterFunc(time.Duration(float64(timeout)*float64(time.Second)), func() {
			timedOut.Store(true)
			thread.Cancel("timed out")
		})
		defer timer.Stop()
	}

	result, err := starlark.Eval(thread, goFilename, goExpr, state.Globals)
	state.ReattachGIL()

	if err != nil {
		if timedOut.Load() {
			raiseTimeoutPythonException(err)
		} else {
			raisePythonException(err)
		}
		return nil
	}

	if convert == 0 {
		cstr := C.CString(result.String())
		defer C.free(unsafe.Pointer(cstr))
		return C.cgoPy_BuildString(cstr)
	} else {
		retval, err := starlarkValueToPython(result)
		if err != nil {
			return nil
		}

		return retval
	}
}

//export Starlark_exec
func Starlark_exec(self *C.Starlark, args *C.PyObject, kwargs *C.PyObject) *C.PyObject {
	var (
		defs       *C.char
		filename   *C.char     = nil
		print      *C.PyObject = nil
		timeout    C.double    = 0
		goFilename string      = "<expr>"
	)

	if C.parseExecArgs(args, kwargs, &defs, &filename, &print, &timeout) == 0 {
		return nil
	}

	print = pythonPrint(self, print)
	if print == nil {
		return nil
	}

	goDefs := C.GoString(defs)

	if filename != nil {
		goFilename = C.GoString(filename)
	}

	state := lockSelf(self)
	if state == nil {
		return nil
	}
	defer state.Mutex.Unlock()

	state.DetachGIL()
	starlarkPrint := func(_ *starlark.Thread, msg string) {
		state.ReattachGIL()
		defer state.DetachGIL()

		callPythonPrint(print, msg)
	}

	_, program, err := starlark.SourceProgram(goFilename, goDefs, state.Globals.Has)
	if err != nil {
		state.ReattachGIL()
		raisePythonException(err)
		return nil
	}

	thread := &starlark.Thread{Print: starlarkPrint}

	var timedOut atomic.Bool
	if timeout > 0 {
		timer := time.AfterFunc(time.Duration(float64(timeout)*float64(time.Second)), func() {
			timedOut.Store(true)
			thread.Cancel("timed out")
		})
		defer timer.Stop()
	}

	newGlobals, err := program.Init(thread, state.Globals)
	state.ReattachGIL()

	if err != nil {
		if timedOut.Load() {
			raiseTimeoutPythonException(err)
		} else {
			raisePythonException(err)
		}
		return nil
	}

	for k, v := range newGlobals {
		v.Freeze()
		state.Globals[k] = v
	}

	return C.cgoPy_NewRef(C.Py_None)
}
