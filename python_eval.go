package main

/*
#include "starlark.h"
*/
import "C"

import (
	"unsafe"

	"go.starlark.net/starlark"
)

//export Starlark_eval
func Starlark_eval(self *C.Starlark, args *C.PyObject, kwargs *C.PyObject) *C.PyObject {
	var (
		expr       *C.char
		filename   *C.char = nil
		convert    C.uint  = 1
		goFilename string  = "<expr>"
	)

	if C.parseEvalArgs(args, kwargs, &expr, &filename, &convert) == 0 {
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

	thread := &starlark.Thread{}
	pyThread := C.PyEval_SaveThread()
	result, err := starlark.Eval(thread, goFilename, goExpr, state.Globals)
	C.PyEval_RestoreThread(pyThread)

	if err != nil {
		raisePythonException(err)
		return nil
	}

	if convert == 0 {
		cstr := C.CString(result.String())
		defer C.free(unsafe.Pointer(cstr))
		return C.cgoPy_BuildString(cstr)
	} else {
		return starlarkValueToPython(result)
	}
}

//export Starlark_exec
func Starlark_exec(self *C.Starlark, args *C.PyObject, kwargs *C.PyObject) *C.PyObject {
	var (
		defs       *C.char
		filename   *C.char = nil
		goFilename string  = "<expr>"
	)

	if C.parseExecArgs(args, kwargs, &defs, &filename) == 0 {
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

	pyThread := C.PyEval_SaveThread()
	_, program, err := starlark.SourceProgram(goFilename, goDefs, state.Globals.Has)
	if err != nil {
		C.PyEval_RestoreThread(pyThread)
		raisePythonException(err)
		return nil
	}

	thread := &starlark.Thread{}
	newGlobals, err := program.Init(thread, state.Globals)
	C.PyEval_RestoreThread(pyThread)

	if err != nil {
		raisePythonException(err)
		return nil
	}

	for k, v := range newGlobals {
		state.Globals[k] = v
	}

	return C.cgoPy_NewRef(C.Py_None)
}
