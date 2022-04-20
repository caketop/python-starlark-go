package main

/*
#include <stdlib.h>
#include <starlark.h>

extern PyObject *StarlarkError;
extern PyObject *SyntaxError;
extern PyObject *EvalError;
extern PyObject *ResolveError;
extern PyObject *ConversionError;
*/
import "C"

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"unsafe"

	"go.starlark.net/resolve"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type StarlarkState struct {
	Globals *starlark.StringDict
	Mutex   *sync.RWMutex
}

var (
	STATE       = map[uint64]*StarlarkState{}
	STATE_MUTEX = sync.Mutex{}
)

func init() {
	resolve.AllowSet = true
}

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

func starlarkIntToPython(x starlark.Int) *C.PyObject {
	/* Try to do it quickly */
	xint, ok := x.Int64()
	if ok {
		return C.PyLong_FromLongLong(C.longlong(xint))
	}

	/* Fall back to converting from string */
	cstr := C.CString(x.String())
	defer C.free(unsafe.Pointer(cstr))
	return C.PyLong_FromString(cstr, nil, 10)
}

func starlarkStringToPython(x starlark.String) *C.PyObject {
	cstr := C.CString(string(x))
	defer C.free(unsafe.Pointer(cstr))
	return C.cgoPy_BuildString(cstr)
}

func starlarkDictToPython(x starlark.IterableMapping) *C.PyObject {
	items := x.Items()
	dict := C.PyDict_New()

	for _, item := range items {
		key := starlarkToPython(item[0])
		defer C.Py_DecRef(key)

		if key == nil {
			C.Py_DecRef(dict)
			return nil
		}

		value := starlarkToPython((item[1]))
		defer C.Py_DecRef(value)

		if value == nil {
			C.Py_DecRef(dict)
			return nil
		}

		// This does not steal references
		C.PyDict_SetItem(dict, key, value)
	}

	return dict
}

func starlarkTupleToPython(x starlark.Tuple) *C.PyObject {
	tuple := C.PyTuple_New(C.Py_ssize_t(x.Len()))
	iter := x.Iterate()
	defer iter.Done()

	var elem starlark.Value
	for i := 0; iter.Next(&elem); i++ {
		value := starlarkToPython(elem)

		if value == nil {
			C.Py_DecRef(value)
			C.Py_DecRef(tuple)
			return nil
		}

		// This "steals" the ref to value so we don't need to DecRef
		C.PyTuple_SetItem(tuple, C.Py_ssize_t(i), value)
	}

	return tuple
}

func starlarkListToPython(x starlark.Iterable) *C.PyObject {
	list := C.PyList_New(0)
	iter := x.Iterate()
	defer iter.Done()

	var elem starlark.Value
	for i := 0; iter.Next(&elem); i++ {
		value := starlarkToPython(elem)

		if value == nil {
			C.Py_DecRef(value)
			C.Py_DecRef(list)
			return nil
		}

		// This "steals" the ref to value so we don't need to DecRef
		C.PyList_Append(list, value)
	}

	return list
}

func starlarkSetToPython(x starlark.Set) *C.PyObject {
	set := C.PySet_New(nil)
	iter := x.Iterate()
	defer iter.Done()

	var elem starlark.Value
	for i := 0; iter.Next(&elem); i++ {
		value := starlarkToPython(elem)
		defer C.Py_DecRef(value)

		if value == nil {
			C.Py_DecRef(set)
			return nil
		}

		// This does not steal references
		C.PySet_Add(set, value)
	}

	return set
}

func starlarkBytesToPython(x starlark.Bytes) *C.PyObject {
	cstr := C.CString(string(x))
	defer C.free(unsafe.Pointer(cstr))
	return C.PyBytes_FromStringAndSize(cstr, C.Py_ssize_t(x.Len()))
}

func starlarkToPython(x starlark.Value) *C.PyObject {
	switch x := x.(type) {
	case starlark.NoneType:
		return C.cgoPy_NewRef(C.Py_None)
	case starlark.Bool:
		if x {
			return C.cgoPy_NewRef(C.Py_True)
		} else {
			return C.cgoPy_NewRef(C.Py_False)
		}
	case starlark.Int:
		return starlarkIntToPython(x)
	case starlark.Float:
		return C.PyFloat_FromDouble(C.double(float64(x)))
	case starlark.String:
		return starlarkStringToPython(x)
	case starlark.Bytes:
		return starlarkBytesToPython(x)
	case *starlark.Set:
		return starlarkSetToPython(*x)
	case starlark.IterableMapping:
		return starlarkDictToPython(x)
	case starlark.Tuple:
		return starlarkTupleToPython(x)
	case starlark.Iterable:
		return starlarkListToPython(x)
	}

	if C.PyErr_Occurred() == nil {
		errmsg := C.CString(fmt.Sprintf("Don't know how to convert %s to Python value", reflect.TypeOf(x).String()))
		defer C.free(unsafe.Pointer(errmsg))
		C.PyErr_SetString(C.ConversionError, errmsg)
	}

	return nil
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

	STATE[stateId] = &StarlarkState{Globals: &starlark.StringDict{}, Mutex: &sync.RWMutex{}}
	self.state_id = C.ulong(stateId)
	return self
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

	stateId := uint64(self.state_id)
	state, ok := STATE[stateId]

	if !ok {
		raiseRuntimeError("Internal error: eval: unknown state_id")
		return nil
	}

	state.Mutex.RLock()
	defer state.Mutex.RUnlock()

	if _, ok = STATE[stateId]; !ok {
		raiseRuntimeError("Internal error: eval: missing state_id after RLock()")
		return nil
	}

	thread := &starlark.Thread{}
	pyThread := C.PyEval_SaveThread()
	result, err := starlark.Eval(thread, goFilename, goExpr, *state.Globals)
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
		return starlarkToPython(result)
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

	stateId := uint64(self.state_id)
	state, ok := STATE[stateId]

	if !ok {
		raiseRuntimeError("Internal error: exec: unknown state_id")
		return nil
	}

	state.Mutex.Lock()
	defer state.Mutex.Unlock()

	if _, ok = STATE[stateId]; !ok {
		raiseRuntimeError("Internal error: exec: missing state_id after Lock()")
		return nil
	}

	thread := &starlark.Thread{}
	pyThread := C.PyEval_SaveThread()
	globals, err := starlark.ExecFile(thread, goFilename, goDefs, *state.Globals)
	C.PyEval_RestoreThread(pyThread)

	if err != nil {
		raisePythonException(err)
		return nil
	}

	state.Globals = &globals
	return C.cgoPy_NewRef(C.Py_None)
}

func main() {}
