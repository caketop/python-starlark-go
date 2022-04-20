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
	Globals starlark.StringDict
	Mutex   sync.RWMutex
}

var (
	STATE       = map[uint64]*StarlarkState{}
	STATE_MUTEX = sync.Mutex{}
)

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

		// This "steals" the ref to value so we don't need to DecRef after
		if C.PyTuple_SetItem(tuple, C.Py_ssize_t(i), value) != 0 {
			C.Py_DecRef(value)
			C.Py_DecRef(tuple)
			return nil
		}
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
			C.Py_DecRef(list)
			return nil
		}

		// This "steals" the ref to value so we don't need to DecRef after
		if C.PyList_Append(list, value) != 0 {
			C.Py_DecRef(value)
			C.Py_DecRef(list)
			return nil
		}
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

func rlockSelf(self *C.Starlark) *StarlarkState {
	stateId := uint64(self.state_id)
	state, ok := STATE[stateId]

	if !ok {
		raiseRuntimeError("Internal error: rlockSelf: unknown state_id")
		return nil
	}

	state.Mutex.RLock()
	return state
}

func lockSelf(self *C.Starlark) *StarlarkState {
	stateId := uint64(self.state_id)
	state, ok := STATE[stateId]

	if !ok {
		raiseRuntimeError("Internal error: lockSelf: unknown state_id")
		return nil
	}

	state.Mutex.Lock()
	return state
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

	self.state_id = C.ulong(stateId)
	STATE[stateId] = &StarlarkState{Globals: starlark.StringDict{}, Mutex: sync.RWMutex{}}
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

//export Starlark_keys
func Starlark_keys(self *C.Starlark, _ *C.PyObject) *C.PyObject {
	state := rlockSelf(self)
	if state == nil {
		return nil
	}
	defer state.Mutex.RUnlock()

	list := C.PyList_New(0)
	for _, key := range state.Globals.Keys() {
		ckey := C.CString(key)
		defer C.free(unsafe.Pointer(ckey))

		pykey := C.cgoPy_BuildString(ckey)
		if pykey == nil {
			C.Py_DecRef(list)
			return nil
		}

		if C.PyList_Append(list, pykey) != 0 {
			C.Py_DecRef(pykey)
			C.Py_DecRef(list)
			return nil
		}
	}

	return list
}

//export Starlark_tp_iter
func Starlark_tp_iter(self *C.Starlark) *C.PyObject {
	keys := Starlark_keys(self, nil)
	if keys == nil {
		return nil
	}
	return C.PyObject_GetIter(keys)
}

//export Starlark_mp_length
func Starlark_mp_length(self *C.Starlark) C.Py_ssize_t {
	state := rlockSelf(self)
	if state == nil {
		return -1
	}
	defer state.Mutex.RUnlock()

	return C.Py_ssize_t(len(state.Globals.Keys()))
}

//export Starlark_mp_subscript
func Starlark_mp_subscript(self *C.Starlark, key *C.PyObject) *C.PyObject {
	keystr := C.PyObject_Str(key)
	if keystr == nil {
		return nil
	}
	defer C.Py_DecRef(keystr)

	var size C.Py_ssize_t
	ckey := C.PyUnicode_AsUTF8AndSize(keystr, &size)
	if ckey == nil {
		return nil
	}

	goKey := C.GoString(ckey)
	state := rlockSelf(self)
	if state == nil {
		return nil
	}
	defer state.Mutex.RUnlock()

	value, ok := state.Globals[goKey]
	if !ok {
		C.PyErr_SetObject(C.PyExc_KeyError, key)
		return nil
	}

	return starlarkToPython(value)
}

/*
func Starlark_mp_ass_subscript(self *C.Starlark, key *C.PyObject, v *C.PyObject) C.int {

}
*/

func main() {}
