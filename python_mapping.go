package main

/*
#include "starlark.h"
*/
import "C"

import (
	"unsafe"
)

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

	return starlarkValueToPython(value)
}

/*
func Starlark_mp_ass_subscript(self *C.Starlark, key *C.PyObject, v *C.PyObject) C.int {

}
*/
