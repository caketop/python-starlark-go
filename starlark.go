package main

/*
#include <stdlib.h>
#include <starlark.h>

extern PyObject *StarlarkError;
extern PyObject *SyntaxError;
extern PyObject *EvalError;
extern PyObject *ResolveError;

extern const char *buildStr;
extern const char *buildUint;
*/
import "C"

import (
	"errors"
	"math/rand"
	"reflect"
	"sync"
	"unsafe"

	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

var (
	THREADS = map[uint64]*starlark.Thread{}
	GLOBALS = map[uint64]starlark.StringDict{}
	MUTEXES = map[uint64]*sync.Mutex{}
)

func raisePythonException(err error) {
	var (
		exc_args   *C.PyObject
		exc_type   *C.PyObject
		syntaxErr  syntax.Error
		evalErr    *starlark.EvalError
		resolveErr resolve.ErrorList
	)

	error_msg := C.CString(err.Error())
	defer C.free(unsafe.Pointer(error_msg))

	error_type := C.CString(reflect.TypeOf(err).String())
	defer C.free(unsafe.Pointer(error_type))

	switch {
	case errors.As(err, &syntaxErr):
		msg := C.CString(syntaxErr.Msg)
		defer C.free(unsafe.Pointer(msg))

		filename := C.CString(syntaxErr.Pos.Filename())
		defer C.free(unsafe.Pointer(filename))

		line := C.uint(syntaxErr.Pos.Line)
		column := C.uint(syntaxErr.Pos.Col)

		exc_args = C.CgoSyntaxErrorArgs(error_msg, error_type, msg, filename, line, column)
		exc_type = C.SyntaxError
	case errors.As(err, &evalErr):
		backtrace := C.CString(evalErr.Backtrace())
		defer C.free(unsafe.Pointer(backtrace))

		exc_args = C.CgoEvalErrorArgs(error_msg, error_type, backtrace)
		exc_type = C.EvalError
	case errors.As(err, &resolveErr):
		items := C.PyTuple_New(C.Py_ssize_t(len(resolveErr)))
		defer C.Py_DecRef(items)

		for i, err := range resolveErr {
			msg := C.CString(err.Msg)
			defer C.free(unsafe.Pointer(msg))

			C.PyTuple_SetItem(items, C.Py_ssize_t(i), C.CgoResolveErrorItem(msg, C.uint(err.Pos.Line), C.uint(err.Pos.Col)))
		}

		exc_args = C.CgoResolveErrorArgs(error_msg, error_type, items)
		exc_type = C.ResolveError
	default:
		exc_args = C.CgoStarlarkErrorArgs(error_msg, error_type)
		exc_type = C.StarlarkError
	}

	C.PyErr_SetObject(exc_type, exc_args)
	C.Py_DecRef(exc_args)
}

//export StarlarkGo_new
func StarlarkGo_new(pytype *C.PyTypeObject, args *C.PyObject, kwargs *C.PyObject) *C.StarlarkGo {
	self := C.CgoStarlarkGoAlloc(pytype)
	if self == nil {
		return nil
	}

	threadId := rand.Uint64()
	thread := &starlark.Thread{}

	THREADS[threadId] = thread
	GLOBALS[threadId] = starlark.StringDict{}
	MUTEXES[threadId] = &sync.Mutex{}

	self.starlark_thread = C.ulong(threadId)
	return self
}

//export StarlarkGo_dealloc
func StarlarkGo_dealloc(self *C.StarlarkGo) {
	threadId := uint64(self.starlark_thread)
	mutex := MUTEXES[threadId]

	mutex.Lock()
	defer mutex.Unlock()

	delete(THREADS, threadId)
	delete(GLOBALS, threadId)
	delete(MUTEXES, threadId)

	C.CgoStarlarkGoDealloc(self)
}

//export StarlarkGo_eval
func StarlarkGo_eval(self *C.StarlarkGo, args *C.PyObject, kwargs *C.PyObject) *C.PyObject {
	var (
		expr       *C.char
		filename   *C.char = nil
		parse      C.uint  = 1
		goFilename string  = "<expr>"
	)

	if C.CgoParseEvalArgs(args, kwargs, &expr, &filename, &parse) == 0 {
		return nil
	}

	goExpr := C.GoString(expr)

	if filename != nil {
		goFilename = C.GoString(filename)
	}

	threadId := uint64(self.starlark_thread)
	mutex := MUTEXES[threadId]

	mutex.Lock()
	defer mutex.Unlock()

	thread := THREADS[threadId]
	globals := GLOBALS[threadId]
	pyThread := C.PyEval_SaveThread()
	result, err := starlark.Eval(thread, goFilename, goExpr, globals)
	C.PyEval_RestoreThread(pyThread)

	if err != nil {
		raisePythonException(err)
		return nil
	}

	cstr := C.CString(result.String())
	defer C.free(unsafe.Pointer(cstr))

	return C.CgoPyBuildOneValue(C.buildStr, unsafe.Pointer(cstr))
}

//export StarlarkGo_exec
func StarlarkGo_exec(self *C.StarlarkGo, args *C.PyObject, kwargs *C.PyObject) *C.PyObject {
	var (
		defs       *C.char
		filename   *C.char = nil
		goFilename string  = "<expr>"
	)

	if C.GgoParseExecArgs(args, kwargs, &defs, &filename) == 0 {
		return nil
	}

	goDefs := C.GoString(defs)

	if filename != nil {
		goFilename = C.GoString(filename)
	}

	threadId := uint64(self.starlark_thread)
	mutex := MUTEXES[threadId]

	mutex.Lock()
	defer mutex.Unlock()

	thread := THREADS[threadId]
	pyThread := C.PyEval_SaveThread()
	globals, err := starlark.ExecFile(thread, goFilename, goDefs, GLOBALS[threadId])
	C.PyEval_RestoreThread(pyThread)

	if err != nil {
		raisePythonException(err)
		return nil
	}

	GLOBALS[threadId] = globals
	return C.CgoPyNone()
}

func main() {}
