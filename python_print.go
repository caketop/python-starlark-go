package main

/*
#include "starlark.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

//export Starlark_get_print
func Starlark_get_print(self *C.Starlark, closure *C.void) *C.PyObject {
	state := rlockSelf(self)
	if state == nil {
		return nil
	}
	defer state.Mutex.RUnlock()

	if state.Print == nil {
		return C.cgoPy_NewRef(C.Py_None)
	}

	return C.cgoPy_NewRef(state.Print)
}

//export Starlark_set_print
func Starlark_set_print(self *C.Starlark, value *C.PyObject, closure *C.void) C.int {
	if value == C.Py_None {
		value = nil
	}

	if value != nil {
		if C.PyCallable_Check(value) != 1 {
			errmsg := C.CString(fmt.Sprintf("%s is not callable", C.GoString(value.ob_type.tp_name)))
			defer C.free(unsafe.Pointer(errmsg))
			C.PyErr_SetString(C.PyExc_TypeError, errmsg)
			return -1
		}
	}

	state := lockSelf(self)
	if state == nil {
		return -1
	}
	defer state.Mutex.Unlock()

	state.Print = C.cgoPy_NewRef(value)
	return 0
}

func pythonPrint(self *C.Starlark, print *C.PyObject) *C.PyObject {
	if print == nil {
		state := rlockSelf(self)
		if state == nil {
			return nil
		}
		defer state.Mutex.RUnlock()
		print = state.Print
	}

	if print == nil {
		print = pythonBuiltinPrint()
	}

	if print == nil {
		errmsg := C.CString("Couldn't find print()?")
		defer C.free(unsafe.Pointer(errmsg))
		C.PyErr_SetString(C.PyExc_TypeError, errmsg)
		return nil
	}

	if C.PyCallable_Check(print) != 1 {
		errmsg := C.CString(fmt.Sprintf("%s is not callable", C.GoString(print.ob_type.tp_name)))
		defer C.free(unsafe.Pointer(errmsg))
		C.PyErr_SetString(C.PyExc_TypeError, errmsg)
		return nil
	}

	return print
}

func pythonBuiltinPrint() *C.PyObject {
	builtins := C.PyEval_GetBuiltins()
	if builtins == nil {
		return nil
	}

	cstr := C.CString("print")
	defer C.free(unsafe.Pointer(cstr))
	return C.PyDict_GetItemString(builtins, cstr)
}

func callPythonPrint(print *C.PyObject, msg string) {
	cmsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cmsg))

	pymsg := C.cgoPy_BuildString(cmsg)
	args := C.PyTuple_New(1)
	defer C.Py_DecRef(args)

	if C.PyTuple_SetItem(args, 0, pymsg) == 0 {
		C.PyObject_CallObject(print, args)
	} else {
		C.Py_DecRef(pymsg)
	}
}
