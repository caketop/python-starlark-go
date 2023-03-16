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
		filename   *C.char     = nil
		convert    C.uint      = 1
		print      *C.PyObject = nil
		goFilename string      = "<expr>"
	)

	if C.parseEvalArgs(args, kwargs, &expr, &filename, &convert, &print) == 0 {
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

	pyThread := C.PyEval_SaveThread()
	starlarkPrint := func(_ *starlark.Thread, msg string) {
		C.PyEval_RestoreThread(pyThread)
		callPythonPrint(print, msg)
		pyThread = C.PyEval_SaveThread()
	}

	thread := &starlark.Thread{Print: starlarkPrint}
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
		goFilename string      = "<expr>"
	)

	if C.parseExecArgs(args, kwargs, &defs, &filename, &print) == 0 {
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

	pyThread := C.PyEval_SaveThread()
	starlarkPrint := func(_ *starlark.Thread, msg string) {
		C.PyEval_RestoreThread(pyThread)
		callPythonPrint(print, msg)
		pyThread = C.PyEval_SaveThread()
	}

	_, program, err := starlark.SourceProgram(goFilename, goDefs, state.Globals.Has)
	if err != nil {
		C.PyEval_RestoreThread(pyThread)
		raisePythonException(err)
		return nil
	}

	thread := &starlark.Thread{Print: starlarkPrint}
	newGlobals, err := program.Init(thread, state.Globals)
	C.PyEval_RestoreThread(pyThread)

	if err != nil {
		raisePythonException(err)
		return nil
	}

	for k, v := range newGlobals {
		v.Freeze()
		state.Globals[k] = v
	}

	return C.cgoPy_NewRef(C.Py_None)
}
