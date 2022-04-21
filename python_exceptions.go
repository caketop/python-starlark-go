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
	"errors"
	"reflect"
	"unsafe"

	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
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

		exc_args = C.makeSyntaxErrorArgs(error_msg, error_type, msg, filename, line, column)
		exc_type = C.SyntaxError
	case errors.As(err, &evalErr):
		backtrace := C.CString(evalErr.Backtrace())
		defer C.free(unsafe.Pointer(backtrace))

		exc_args = C.makeEvalErrorArgs(error_msg, error_type, backtrace)
		exc_type = C.EvalError
	case errors.As(err, &resolveErr):
		items := C.PyTuple_New(C.Py_ssize_t(len(resolveErr)))
		defer C.Py_DecRef(items)

		for i, err := range resolveErr {
			msg := C.CString(err.Msg)
			defer C.free(unsafe.Pointer(msg))

			C.PyTuple_SetItem(items, C.Py_ssize_t(i), C.makeResolveErrorItem(msg, C.uint(err.Pos.Line), C.uint(err.Pos.Col)))
		}

		exc_args = C.makeResolveErrorArgs(error_msg, error_type, items)
		exc_type = C.ResolveError
	default:
		exc_args = C.makeStarlarkErrorArgs(error_msg, error_type)
		exc_type = C.StarlarkError
	}

	C.PyErr_SetObject(exc_type, exc_args)
	C.Py_DecRef(exc_args)
}

func raiseRuntimeError(msg string) {
	cmsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cmsg))
	C.PyErr_SetString(C.PyExc_RuntimeError, cmsg)
}
