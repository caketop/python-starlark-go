package main

/*
#include <stdlib.h>
#include <starlark.h>
*/
import "C"

import (
	"errors"
	"math/rand"
	"reflect"
	"unsafe"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

var THREADS = map[uint64]*starlark.Thread{}
var GLOBALS = map[uint64]starlark.StringDict{}

func newThread() C.ulong {
	threadId := rand.Uint64()
	thread := &starlark.Thread{}
	THREADS[threadId] = thread
	GLOBALS[threadId] = starlark.StringDict{}
	return C.ulong(threadId)
}

func destroyThread(threadId C.ulong) {
	goThreadId := uint64(threadId)
	delete(THREADS, goThreadId)
	delete(GLOBALS, goThreadId)
}

func raisePythonException(err error) {
	var exc_args *C.PyObject
	var exc_type *C.PyObject
	var syntaxErr syntax.Error
	var evalErr *starlark.EvalError = nil

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
	default:
		exc_args = C.CgoStarlarkErrorArgs(error_msg, error_type)
		exc_type = C.StarlarkError
	}

	C.PyErr_SetObject(exc_type, exc_args)
	C.CgoPyDecRef(exc_args)
}

//export StarlarkGo_new
func StarlarkGo_new(pytype *C.PyTypeObject, args *C.PyObject, kwargs *C.PyObject) *C.StarlarkGo {
	self := C.CgoStarlarkGoAlloc(pytype)
	if self != nil {
		self.starlark_thread = newThread()
	}
	return self
}

//export StarlarkGo_dealloc
func StarlarkGo_dealloc(self *C.StarlarkGo) {
	destroyThread(self.starlark_thread)
	C.CgoStarlarkGoDealloc(self)
}

//export StarlarkGo_eval
func StarlarkGo_eval(self *C.StarlarkGo, args *C.PyObject) *C.PyObject {
	stmt := C.CgoParseEvalArgs(args)
	if stmt == nil {
		return nil
	}

	defer C.CgoPyDecRef(stmt)

	goStmt := C.GoString(C.PyBytes_AsString(stmt))
	goThreadId := uint64(self.starlark_thread)

	thread := THREADS[goThreadId]
	globals := GLOBALS[goThreadId]

	result, err := starlark.Eval(thread, "<expr>", goStmt, globals)
	if err != nil {
		raisePythonException(err)
		return nil
	}

	cstr := C.CString(result.String())
	retval := C.CgoPyString(cstr)
	C.free(unsafe.Pointer(cstr))

	return retval
}

//export StarlarkGo_exec
func StarlarkGo_exec(self *C.StarlarkGo, args *C.PyObject) *C.PyObject {
	stmt := C.CgoParseEvalArgs(args)
	if stmt == nil {
		return nil
	}

	defer C.CgoPyDecRef(stmt)

	goStmt := C.GoString(C.PyBytes_AsString(stmt))
	goThreadId := uint64(self.starlark_thread)

	thread := THREADS[goThreadId]
	globals, err := starlark.ExecFile(thread, "main.star", goStmt, starlark.StringDict{})

	if err != nil {
		raisePythonException(err)
		return nil
	}

	GLOBALS[goThreadId] = globals

	return C.CgoPyNone()
}

func main() {}
