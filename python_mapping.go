package main

/*
#include "starlark.h"

extern PyObject *ConversionError;
*/
import "C"

import (
	"fmt"
	"math/big"
	"unsafe"

	"go.starlark.net/starlark"
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

//export Starlark_mp_ass_subscript
func Starlark_mp_ass_subscript(self *C.Starlark, key *C.PyObject, v *C.PyObject) C.int {
	keystr := C.PyObject_Str(key)
	if keystr == nil {
		return -1
	}
	defer C.Py_DecRef(keystr)

	var size C.Py_ssize_t
	ckey := C.PyUnicode_AsUTF8AndSize(keystr, &size)
	if ckey == nil {
		return -1
	}
	goKey := C.GoString(ckey)

	var value starlark.Value = nil
	var err error = nil
	if v != nil {
		value, err = pythonToStarlarkValue(v)
		if err != nil {
			return -1
		}
	}

	state := rlockSelf(self)
	if state == nil {
		return -1
	}
	defer state.Mutex.RUnlock()

	if value == nil {
		delete(state.Globals, goKey)
	} else {
		state.Globals[goKey] = value
	}

	return 0
}

func pythonToStarlarkTuple(obj *C.PyObject) (starlark.Tuple, error) {
	var elems []starlark.Value
	pyiter := C.PyObject_GetIter(obj)
	if pyiter == nil {
		return starlark.Tuple{}, fmt.Errorf("List: couldn't get iterator")
	}
	defer C.Py_DecRef(pyiter)

	for pyvalue := C.PyIter_Next(pyiter); pyvalue != nil; pyvalue = C.PyIter_Next(pyiter) {
		defer C.Py_DecRef(pyvalue)

		value, err := pythonToStarlarkValue(pyvalue)
		if err != nil {
			return starlark.Tuple{}, err
		}

		elems = append(elems, value)
	}

	return starlark.Tuple(elems), nil
}

func pythonToStarlarkBytes(obj *C.PyObject) (starlark.Bytes, error) {
	cbytes := C.PyBytes_AsString(obj)
	if cbytes == nil {
		return starlark.Bytes(""), fmt.Errorf("Bytes: couldn't get pointer")
	}

	return starlark.Bytes(C.GoString(cbytes)), nil
}

func pythonToStarlarkList(obj *C.PyObject) (*starlark.List, error) {
	len := C.PyObject_Length(obj)
	if len < 0 {
		return &starlark.List{}, fmt.Errorf("List: couldn't get length of object")
	}

	var elems []starlark.Value
	pyiter := C.PyObject_GetIter(obj)
	if pyiter == nil {
		return &starlark.List{}, fmt.Errorf("List: couldn't get iterator")
	}
	defer C.Py_DecRef(pyiter)

	for pyvalue := C.PyIter_Next(pyiter); pyvalue != nil; pyvalue = C.PyIter_Next(pyiter) {
		defer C.Py_DecRef(pyvalue)

		value, err := pythonToStarlarkValue(pyvalue)
		if err != nil {
			return &starlark.List{}, err
		}

		elems = append(elems, value)
	}

	return starlark.NewList(elems), nil
}

func pythonToStarlarkDict(obj *C.PyObject) (*starlark.Dict, error) {
	len := C.PyObject_Length(obj)
	if len < 0 {
		return &starlark.Dict{}, fmt.Errorf("Dict: couldn't get length of object")
	}

	dict := starlark.NewDict(int(len))
	pyiter := C.PyObject_GetIter(obj)
	if pyiter == nil {
		return &starlark.Dict{}, fmt.Errorf("Dict: couldn't get iterator")
	}
	defer C.Py_DecRef(pyiter)

	for pykey := C.PyIter_Next(pyiter); pykey != nil; pykey = C.PyIter_Next(pyiter) {
		defer C.Py_DecRef(pykey)

		pyvalue := C.PyObject_GetItem(obj, pykey)
		if pyvalue == nil {
			return &starlark.Dict{}, fmt.Errorf("Dict: couldn't get value")
		}
		defer C.Py_DecRef(pyvalue)

		key, err := pythonToStarlarkValue(pykey)
		if err != nil {
			return &starlark.Dict{}, err
		}

		value, err := pythonToStarlarkValue(pyvalue)
		if err != nil {
			return &starlark.Dict{}, err
		}

		err = dict.SetKey(key, value)
		if err != nil {
			return &starlark.Dict{}, err
		}
	}

	return dict, nil
}

func pythonToStarlarkSet(obj *C.PyObject) (*starlark.Set, error) {
	len := C.PyObject_Length(obj)
	if len < 0 {
		return &starlark.Set{}, fmt.Errorf("Set: couldn't get length of object")
	}

	set := starlark.NewSet(int(len))
	pyiter := C.PyObject_GetIter(obj)
	if pyiter == nil {
		return &starlark.Set{}, fmt.Errorf("Set: couldn't get iterator")
	}
	defer C.Py_DecRef(pyiter)

	for pyvalue := C.PyIter_Next(pyiter); pyvalue != nil; pyvalue = C.PyIter_Next(pyiter) {
		defer C.Py_DecRef(pyvalue)

		value, err := pythonToStarlarkValue(pyvalue)
		if err != nil {
			return &starlark.Set{}, err
		}

		err = set.Insert(value)
		if err != nil {
			raisePythonException(err)
			return &starlark.Set{}, err
		}
	}

	return set, nil
}

func pythonToStarlarkString(obj *C.PyObject) (starlark.String, error) {
	var size C.Py_ssize_t
	cstr := C.PyUnicode_AsUTF8AndSize(obj, &size)
	if cstr == nil {
		return starlark.String(""), fmt.Errorf("Int: couldn't convert to C string")
	}

	return starlark.String(C.GoString(cstr)), nil
}

func pythonToStarlarkInt(obj *C.PyObject) (starlark.Int, error) {
	overflow := C.int(0)
	longlong := int64(C.PyLong_AsLongLongAndOverflow(obj, &overflow)) // https://youtu.be/6-1Ue0FFrHY
	if C.PyErr_Occurred() != nil {
		return starlark.Int{}, fmt.Errorf("Int: couldn't convert to long")
	}

	if overflow == 0 {
		return starlark.MakeInt64(longlong), nil
	}

	pystr := C.PyObject_Str(obj)
	if pystr == nil {
		return starlark.Int{}, fmt.Errorf("Int: couldn't convert to Python string")
	}
	defer C.Py_DecRef(pystr)

	var size C.Py_ssize_t
	cstr := C.PyUnicode_AsUTF8AndSize(pystr, &size)
	if cstr == nil {
		return starlark.Int{}, fmt.Errorf("Int: couldn't convert to C string")
	}

	i := new(big.Int)
	i.SetString(C.GoString(cstr), 10)
	return starlark.MakeBigInt(i), nil
}

func pythonToStarlarkFloat(obj *C.PyObject) (starlark.Float, error) {
	cvalue := C.PyFloat_AsDouble(obj)
	if C.PyErr_Occurred() != nil {
		return starlark.Float(0), fmt.Errorf("Float: couldn't conver to double")
	}

	return starlark.Float(cvalue), nil
}

func pythonToStarlarkValue(obj *C.PyObject) (starlark.Value, error) {
	var value starlark.Value = nil
	var err error = nil

	switch {
	case obj == C.Py_None:
		value = starlark.None
	case obj == C.Py_True:
		value = starlark.True
	case obj == C.Py_False:
		value = starlark.False
	case C.cgoPyFloat_Check(obj) == 1:
		value, err = pythonToStarlarkFloat(obj)
	case C.cgoPyLong_Check(obj) == 1:
		value, err = pythonToStarlarkInt(obj)
	case C.cgoPyUnicode_Check(obj) == 1:
		value, err = pythonToStarlarkString(obj)
	case C.cgoPyBytes_Check(obj) == 1:
		value, err = pythonToStarlarkBytes(obj)
	case C.cgoPySet_Check(obj) == 1:
		value, err = pythonToStarlarkSet(obj)
	case C.cgoPyDict_Check(obj) == 1:
		value, err = pythonToStarlarkDict(obj)
	case C.cgoPyList_Check(obj) == 1:
		value, err = pythonToStarlarkList(obj)
	case C.cgoPyTuple_Check(obj) == 1:
		value, err = pythonToStarlarkTuple(obj)
	case C.PyMapping_Check(obj) == 1:
		value, err = pythonToStarlarkDict(obj)
	case C.PySequence_Check(obj) == 1:
		value, err = pythonToStarlarkList(obj)
	}

	if err != nil {
		if C.PyErr_Occurred() == nil {
			var errmsg *C.char
			tp_name := C.GoString(obj.ob_type.tp_name)

			if value == nil {
				errmsg = C.CString(fmt.Sprintf("Don't know how to convert %s to Starlark value", tp_name))
			} else {
				errmsg = C.CString(fmt.Sprintf("While converting %s to Starlark value: %s", tp_name, err.Error()))
			}
			defer C.free(unsafe.Pointer(errmsg))

			C.PyErr_SetString(C.ConversionError, errmsg)
		}
	}

	return value, err
}
