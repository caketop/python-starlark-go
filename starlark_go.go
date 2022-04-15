package main

/*
#include <stdlib.h>
void Raise_StarlarkError(const char *error, const char *error_type);
void Raise_SyntaxError(const char *error, const char *error_type, const char *msg, const char *filename, const long line, const long column);
void Raise_EvalError(const char *error, const char *error_type, const char *backtrace);
*/
import "C"

import (
	"encoding/json"
	"fmt"
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

func raiseStarlarkError(err error) {
	error_str := C.CString(err.Error())
	error_type := C.CString(fmt.Sprintf("%s", reflect.TypeOf(err)))

	C.Raise_StarlarkError(error_str, error_type)

	FreeCString(error_type)
	FreeCString(error_str)
}


func raiseSyntaxError(err *syntax.Error) {
	error_str := C.CString(err.Error())
	error_type := C.CString(fmt.Sprintf("%s", reflect.TypeOf(err)))
	msg := C.CString(err.Msg)
	filename := C.CString(err.Pos.Filename())

	C.Raise_SyntaxError(error_str, error_type, msg, filename, C.long(err.Pos.Line), C.long(err.Pos.Col))

	FreeCString(filename)
	FreeCString(msg)
	FreeCString(error_type)
	FreeCString(error_str)
}

func raiseEvalError(err *starlark.EvalError) {
	error_str := C.CString(err.Error())
	error_type := C.CString(fmt.Sprintf("%s", reflect.TypeOf(err)))
	backtrace := C.CString(err.Backtrace())

	C.Raise_EvalError(error_str, error_type, backtrace)

	FreeCString(backtrace)
	FreeCString(error_type)
	FreeCString(error_str)
}

func handleError(err error) {
	syntaxErr, ok := err.(syntax.Error)
	if ok {
		raiseSyntaxError(&syntaxErr)
		return
	}

	evalErr, ok := err.(*starlark.EvalError)
	if ok {
		raiseEvalError(evalErr)
		return
	}

	raiseStarlarkError(err)
}

//export Eval
func Eval(threadId C.ulong, stmt *C.char) *C.char {
	// Cast *char to string
	goStmt := C.GoString(stmt)
	goThreadId := uint64(threadId)

	thread := THREADS[goThreadId]
	globals := GLOBALS[goThreadId]

	result, err := starlark.Eval(thread, "<expr>", goStmt, globals)
	if err != nil {
		handleError(err)
		return nil
	}

	// Convert starlark.Value struct into a JSON blob
	rawResponse := make(map[string]string)
	rawResponse["value"] = result.String()
	rawResponse["type"] = result.Type()
	response, _ := json.Marshal(rawResponse)

	// Convert JSON blob to string and then CString
	return C.CString(string(response))
}

//export ExecFile
func ExecFile(threadId C.ulong, data *C.char) C.int {
	// Cast *char to string
	goData := C.GoString(data)
	goThreadId := uint64(threadId)

	thread := THREADS[goThreadId]
	globals, err := starlark.ExecFile(thread, "main.star", goData, starlark.StringDict{})
	if err != nil {
		handleError(err)
		return C.int(0)
	}
	GLOBALS[goThreadId] = globals
	return C.int(1)
}

//export FreeCString
func FreeCString(s *C.char) {
	C.free(unsafe.Pointer(s))
}

func main() {}

/*
func main() {
	const data = `
def fibonacci(n=10):
	res = list(range(n))
	for i in res[2:]:
		res[i] = res[i-2] + res[i-1]
	return res
`
	threadId := NewThread()
	ExecFile(threadId, C.CString(data))
	r := Eval(threadId, C.CString("fibonacci(25)"))
	fmt.Printf("%v\n", C.GoString(r))
	DestroyThread(threadId)
}
*/
