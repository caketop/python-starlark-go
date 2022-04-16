package main

/*
#include <stdlib.h>
#include <starlark.h>
*/
import "C"

import (
	"math/rand"
	"reflect"
	"unsafe"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

var THREADS = map[uint64]*starlark.Thread{}
var GLOBALS = map[uint64]starlark.StringDict{}

//export NewThread
func NewThread() C.ulong {
	threadId := rand.Uint64()
	thread := &starlark.Thread{}
	THREADS[threadId] = thread
	GLOBALS[threadId] = starlark.StringDict{}
	return C.ulong(threadId)
}

//export DestroyThread
func DestroyThread(threadId C.ulong) {
	goThreadId := uint64(threadId)
	delete(THREADS, goThreadId)
	delete(GLOBALS, goThreadId)
}

func makeStarlarkErrorArgs(err error) *C.StarlarkErrorArgs {
	args := (*C.StarlarkErrorArgs)(C.malloc(C.sizeof_StarlarkErrorArgs))
	args.error = C.CString(err.Error())
	args.error_type = C.CString(reflect.TypeOf(err).String())

	return args
}

func makeSyntaxErrorArgs(err *syntax.Error) *C.SyntaxErrorArgs {
	args := (*C.SyntaxErrorArgs)(C.malloc(C.sizeof_SyntaxErrorArgs))
	args.error = C.CString(err.Error())
	args.error_type = C.CString(reflect.TypeOf(err).String())
	args.msg = C.CString(err.Msg)
	args.filename = C.CString(err.Pos.Filename())
	args.line = C.uint(err.Pos.Line)
	args.column = C.uint(err.Pos.Col)

	return args
}

func makeEvalErrorArgs(err *starlark.EvalError) *C.EvalErrorArgs {
	args := (*C.EvalErrorArgs)(C.malloc(C.sizeof_EvalErrorArgs))
	args.error = C.CString(err.Error())
	args.error_type = C.CString(reflect.TypeOf(err).String())
	args.backtrace = C.CString(err.Backtrace())

	return args
}

func makeStarlarkReturn(err error) *C.StarlarkReturn {
	retval := (*C.StarlarkReturn)(C.malloc(C.sizeof_StarlarkReturn))
	retval.value = nil

	if err != nil {
		syntaxErr, ok := err.(syntax.Error)
		if ok {
			retval.error_type = C.STARLARK_SYNTAX_ERROR
			retval.error = unsafe.Pointer(makeSyntaxErrorArgs(&syntaxErr))
			return retval
		}

		evalErr, ok := err.(*starlark.EvalError)
		if ok {
			retval.error_type = C.STARLARK_EVAL_ERROR
			retval.error = unsafe.Pointer(makeEvalErrorArgs(evalErr))
			return retval
		}

		retval.error_type = C.STARLARK_GENERAL_ERROR
		retval.error = unsafe.Pointer(makeStarlarkErrorArgs(err))
		return retval
	}

	retval.error_type = C.STARLARK_NO_ERROR
	retval.error = nil

	return retval
}

//export FreeStarlarkReturn
func FreeStarlarkReturn(retval *C.StarlarkReturn) {
	switch retval.error_type {
	case C.STARLARK_GENERAL_ERROR:
		args := (*C.StarlarkErrorArgs)(retval.error)
		C.free(unsafe.Pointer(args.error))
		C.free(unsafe.Pointer(args.error_type))
	case C.STARLARK_SYNTAX_ERROR:
		args := (*C.SyntaxErrorArgs)(retval.error)
		C.free(unsafe.Pointer(args.error))
		C.free(unsafe.Pointer(args.error_type))
		C.free(unsafe.Pointer(args.msg))
		C.free(unsafe.Pointer(args.filename))
	case C.STARLARK_EVAL_ERROR:
		args := (*C.EvalErrorArgs)(retval.error)
		C.free(unsafe.Pointer(args.error))
		C.free(unsafe.Pointer(args.error_type))
		C.free(unsafe.Pointer(args.backtrace))
	case C.STARLARK_NO_ERROR:
		if retval.error != nil {
			panic("STARLARK_NO_ERROR but error is not nil")
		}
	default:
		panic("unknown error_type")
	}

	if retval.value != nil {
		C.free(unsafe.Pointer(retval.value))
	}

	if retval.error != nil {
		C.free(unsafe.Pointer(retval.error))
	}

	C.free(unsafe.Pointer(retval))
}

//export Eval
func Eval(threadId C.ulong, stmt *C.char) *C.StarlarkReturn {
	// Cast *char to string
	goStmt := C.GoString(stmt)
	goThreadId := uint64(threadId)

	thread := THREADS[goThreadId]
	globals := GLOBALS[goThreadId]

	result, err := starlark.Eval(thread, "<expr>", goStmt, globals)
	retval := makeStarlarkReturn(err)

	if err == nil {
		retval.value = C.CString(result.String())
	}

	return retval
}

//export ExecFile
func ExecFile(threadId C.ulong, data *C.char) *C.StarlarkReturn {
	// Cast *char to string
	goData := C.GoString(data)
	goThreadId := uint64(threadId)

	thread := THREADS[goThreadId]
	globals, err := starlark.ExecFile(thread, "main.star", goData, starlark.StringDict{})
	retval := makeStarlarkReturn(err)

	if err == nil {
		GLOBALS[goThreadId] = globals
	}

	return retval
}

func main() {}
