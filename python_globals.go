package main

/*
#include "starlark.h"

extern PyObject *ConversionError;
*/
import "C"

import (
	"unsafe"
)

//export Starlark_global_names
func Starlark_global_names(self *C.Starlark, _ *C.PyObject) *C.PyObject {
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

//export Starlark_get_global
func Starlark_get_global(self *C.Starlark, args *C.PyObject, kwargs *C.PyObject) *C.PyObject {
	var name *C.char = nil
	var default_value *C.PyObject = nil

	if C.parseGetGlobalArgs(args, kwargs, &name, &default_value) == 0 {
		return nil
	}

	goName := C.GoString(name)
	state := rlockSelf(self)
	if state == nil {
		return nil
	}
	defer state.Mutex.RUnlock()

	value, ok := state.Globals[goName]
	if !ok {
		if default_value != nil {
			return default_value
		}

		C.PyErr_SetString(C.PyExc_KeyError, name)
		return nil
	}

	retval, err := starlarkValueToPython(value)
	if err != nil {
		return nil
	}

	return retval
}

//export Starlark_set_globals
func Starlark_set_globals(self *C.Starlark, args *C.PyObject, kwargs *C.PyObject) *C.PyObject {
	posargs := C.PyObject_Length(args)

	if posargs > 0 {
		errmsg := C.CString("set_globals takes no positional arguments")
		defer C.free(unsafe.Pointer(errmsg))
		C.PyErr_SetString(C.PyExc_TypeError, errmsg)
		return nil
	}

	if kwargs == nil {
		return C.cgoPy_NewRef(C.Py_None)
	}

	pyiter := C.PyObject_GetIter(kwargs)
	if pyiter == nil {
		return nil
	}
	defer C.Py_DecRef(pyiter)

	state := rlockSelf(self)
	if state == nil {
		return nil
	}
	defer state.Mutex.RUnlock()

	for pykey := C.PyIter_Next(pyiter); pykey != nil; pykey = C.PyIter_Next(pyiter) {
		defer C.Py_DecRef(pykey)

		var size C.Py_ssize_t
		ckey := C.PyUnicode_AsUTF8AndSize(pykey, &size)
		if ckey == nil {
			return nil
		}
		key := C.GoString(ckey)

		pyvalue := C.PyObject_GetItem(kwargs, pykey)
		if pyvalue == nil {
			return nil
		}
		defer C.Py_DecRef(pyvalue)

		value, err := pythonToStarlarkValue(pyvalue)
		if err != nil {
			return nil
		}

		value.Freeze()
		state.Globals[key] = value
	}

	return C.cgoPy_NewRef(C.Py_None)
}

//export Starlark_pop_global
func Starlark_pop_global(self *C.Starlark, args *C.PyObject, kwargs *C.PyObject) *C.PyObject {
	var name *C.char = nil
	var default_value *C.PyObject = nil

	if C.parsePopGlobalArgs(args, kwargs, &name, &default_value) == 0 {
		return nil
	}

	goName := C.GoString(name)
	state := rlockSelf(self)
	if state == nil {
		return nil
	}
	defer state.Mutex.RUnlock()

	value, ok := state.Globals[goName]
	if !ok {
		if default_value != nil {
			return default_value
		}

		C.PyErr_SetString(C.PyExc_KeyError, name)
		return nil
	}

	delete(state.Globals, goName)
	retval, err := starlarkValueToPython(value)
	if err != nil {
		return nil
	}

	return retval
}

//export Starlark_tp_iter
func Starlark_tp_iter(self *C.Starlark) *C.PyObject {
	keys := Starlark_global_names(self, nil)
	if keys == nil {
		return nil
	}
	return C.PyObject_GetIter(keys)
}
