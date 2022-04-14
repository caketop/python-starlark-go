package main

/*
#include <stdlib.h>
void Raise_EvalError(const char *message);
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"math/rand"
	"unsafe"

	"go.starlark.net/starlark"
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

//export Eval
func Eval(threadId C.ulong, stmt *C.char) *C.char {
	// Cast *char to string
	goStmt := C.GoString(stmt)
	goThreadId := uint64(threadId)

	thread := THREADS[goThreadId]
	globals := GLOBALS[goThreadId]

	result, err := starlark.Eval(thread, "<expr>", goStmt, globals)
	if err != nil {
		message := C.CString(fmt.Sprintf("%v", err))
		C.Raise_EvalError(message)
		FreeCString(message)
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
		message := C.CString(fmt.Sprintf("%v", err))
		C.Raise_EvalError(message)
		FreeCString(message)
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
