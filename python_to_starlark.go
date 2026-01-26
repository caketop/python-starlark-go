package main

/*
#include "starlark.h"

extern PyObject *ConversionToStarlarkFailed;
*/
import "C"

import (
	"fmt"
	"math/big"
	"unsafe"

	"go.starlark.net/starlark"
)

func (state *StarlarkState) pythonToStarlarkTuple(obj *C.PyObject) (starlark.Tuple, error) {
	var elems []starlark.Value
	pyiter := C.PyObject_GetIter(obj)
	if pyiter == nil {
		return starlark.Tuple{}, fmt.Errorf("Couldn't get iterator for Python tuple")
	}
	defer C.Py_DecRef(pyiter)

	index := 0
	for pyvalue := C.PyIter_Next(pyiter); pyvalue != nil; pyvalue = C.PyIter_Next(pyiter) {
		defer C.Py_DecRef(pyvalue)

		value, err := state.innerPythonToStarlarkValue(pyvalue)
		if err != nil {
			return starlark.Tuple{}, fmt.Errorf("While converting value at index %v in Python tuple: %v", index, err)
		}

		elems = append(elems, value)
		index += 1
	}

	if C.PyErr_Occurred() != nil {
		return starlark.Tuple{}, fmt.Errorf("Python exception while converting value at index %v in Python tuple", index)
	}

	return starlark.Tuple(elems), nil
}

func pythonToStarlarkBytes(obj *C.PyObject) (starlark.Bytes, error) {
	cbytes := C.PyBytes_AsString(obj)
	if cbytes == nil {
		return starlark.Bytes(""), fmt.Errorf("Couldn't get pointer to Python bytes")
	}

	return starlark.Bytes(C.GoString(cbytes)), nil
}

func (state *StarlarkState) pythonToStarlarkList(obj *C.PyObject) (*starlark.List, error) {
	len := C.PyObject_Length(obj)
	if len < 0 {
		return &starlark.List{}, fmt.Errorf("Couldn't get size of Python list")
	}

	var elems []starlark.Value
	pyiter := C.PyObject_GetIter(obj)
	if pyiter == nil {
		return &starlark.List{}, fmt.Errorf("Couldn't get iterator for Python list")
	}
	defer C.Py_DecRef(pyiter)

	index := 0
	for pyvalue := C.PyIter_Next(pyiter); pyvalue != nil; pyvalue = C.PyIter_Next(pyiter) {
		defer C.Py_DecRef(pyvalue)
		value, err := state.innerPythonToStarlarkValue(pyvalue)
		if err != nil {
			return &starlark.List{}, fmt.Errorf("While converting value at index %v in Python list: %v", index, err)
		}

		elems = append(elems, value)
		index += 1
	}

	if C.PyErr_Occurred() != nil {
		return &starlark.List{}, fmt.Errorf("Python exception while converting value at index %v in Python list", index)
	}

	return starlark.NewList(elems), nil
}

func (state *StarlarkState) pythonToStarlarkDict(obj *C.PyObject) (*starlark.Dict, error) {
	size := C.PyObject_Length(obj)
	if size < 0 {
		return &starlark.Dict{}, fmt.Errorf("Couldn't get size of Python dict")
	}

	dict := starlark.NewDict(int(size))
	pyiter := C.PyObject_GetIter(obj)
	if pyiter == nil {
		return &starlark.Dict{}, fmt.Errorf("Couldn't get iterator for Python dict")
	}
	defer C.Py_DecRef(pyiter)

	for pykey := C.PyIter_Next(pyiter); pykey != nil; pykey = C.PyIter_Next(pyiter) {
		defer C.Py_DecRef(pykey)

		key, err := state.innerPythonToStarlarkValue(pykey)
		if err != nil {
			return &starlark.Dict{}, fmt.Errorf("While converting key in Python dict: %v", err)
		}

		pyvalue := C.PyObject_GetItem(obj, pykey)
		if pyvalue == nil {
			return &starlark.Dict{}, fmt.Errorf("Couldn't get value of key %v in Python dict", key)
		}
		defer C.Py_DecRef(pyvalue)

		value, err := state.innerPythonToStarlarkValue(pyvalue)
		if err != nil {
			return &starlark.Dict{}, fmt.Errorf("While converting value of key %v in Python dict: %v", key, err)
		}

		err = dict.SetKey(key, value)
		if err != nil {
			return &starlark.Dict{}, fmt.Errorf("While setting %v to %v in Starlark dict: %v", key, value, err)
		}
	}

	if C.PyErr_Occurred() != nil {
		return &starlark.Dict{}, fmt.Errorf("Python exception while iterating through Python dict")
	}

	return dict, nil
}

func (state *StarlarkState) pythonToStarlarkSet(obj *C.PyObject) (*starlark.Set, error) {
	size := C.PyObject_Length(obj)
	if size < 0 {
		return &starlark.Set{}, fmt.Errorf("Couldn't get size of Python set")
	}

	set := starlark.NewSet(int(size))
	pyiter := C.PyObject_GetIter(obj)
	if pyiter == nil {
		return &starlark.Set{}, fmt.Errorf("Couldn't get iterator for Python set")
	}
	defer C.Py_DecRef(pyiter)

	for pyvalue := C.PyIter_Next(pyiter); pyvalue != nil; pyvalue = C.PyIter_Next(pyiter) {
		defer C.Py_DecRef(pyvalue)

		value, err := state.innerPythonToStarlarkValue(pyvalue)
		if err != nil {
			return &starlark.Set{}, fmt.Errorf("While converting value in Python set: %v", err)
		}

		err = set.Insert(value)
		if err != nil {
			raisePythonException(err)
			return &starlark.Set{}, fmt.Errorf("While inserting %v into Starlark set: %v", value, err)
		}
	}

	if C.PyErr_Occurred() != nil {
		return &starlark.Set{}, fmt.Errorf("Python exception while converting value in Python set to Starlark")
	}

	return set, nil
}

func pythonToStarlarkString(obj *C.PyObject) (starlark.String, error) {
	var size C.Py_ssize_t
	cstr := C.PyUnicode_AsUTF8AndSize(obj, &size)
	if cstr == nil {
		return starlark.String(""), fmt.Errorf("Couldn't convert Python string to C string")
	}

	return starlark.String(C.GoString(cstr)), nil
}

func pythonToStarlarkInt(obj *C.PyObject) (starlark.Int, error) {
	overflow := C.int(0)
	longlong := int64(C.PyLong_AsLongLongAndOverflow(obj, &overflow)) // https://youtu.be/6-1Ue0FFrHY
	if C.PyErr_Occurred() != nil {
		return starlark.Int{}, fmt.Errorf("Couldn't convert Python int to long")
	}

	if overflow == 0 {
		return starlark.MakeInt64(longlong), nil
	}

	pystr := C.PyObject_Str(obj)
	if pystr == nil {
		return starlark.Int{}, fmt.Errorf("Couldn't convert Python int to string")
	}
	defer C.Py_DecRef(pystr)

	var size C.Py_ssize_t
	cstr := C.PyUnicode_AsUTF8AndSize(pystr, &size)
	if cstr == nil {
		return starlark.Int{}, fmt.Errorf("Couldn't convert Python int to C string")
	}

	i := new(big.Int)
	i.SetString(C.GoString(cstr), 10)
	return starlark.MakeBigInt(i), nil
}

func pythonToStarlarkFloat(obj *C.PyObject) (starlark.Float, error) {
	cvalue := C.PyFloat_AsDouble(obj)
	if C.PyErr_Occurred() != nil {
		return starlark.Float(0), fmt.Errorf("Couldn't convert Python float to double")
	}

	return starlark.Float(cvalue), nil
}

func getFuncName(obj *C.PyObject) (string, error) {
	nameAttr := C.CString("__name__")
	defer C.free(unsafe.Pointer(nameAttr))
	funcName, err := pythonToStarlarkString(C.PyObject_GetAttrString(obj, nameAttr))
	if err != nil {
		return "", err
	}
	return funcName.GoString(), nil
}

func getPyError() error {
	// TODO: replace with PyErr_GetRaisedException when requiring Python >= 3.12
	var (
		errType      *C.PyObject
		errValue     *C.PyObject
		errTraceback *C.PyObject
	)
	C.PyErr_Fetch(&errType, &errValue, &errTraceback)
	defer C.Py_DecRef(errType)
	defer C.Py_DecRef(errValue)
	defer C.Py_DecRef(errTraceback)

	errStr, err := pythonToStarlarkString(C.PyObject_Str(errValue))
	if err != nil {
		return err
	}

	return fmt.Errorf(errStr.GoString())
}

func (state *StarlarkState) pythonToStarlarkFunc(obj *C.PyObject) (starlark.Value, error) {
	funcName, err := getFuncName(obj)
	if err != nil {
		return starlark.None, err
	}

	C.Py_IncRef(obj)
	state.childRefs = append(state.childRefs, obj)

	return starlark.NewBuiltin(funcName, func(
		_ *starlark.Thread,
		_ *starlark.Builtin,
		args starlark.Tuple,
		kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		state.ReattachGIL()
		defer state.DetachGIL()

		cargs, err := starlarkTupleToPython(args)
		if err != nil {
			return starlark.None, err
		}
		defer C.Py_DecRef(cargs)

		ckwargs, err := starlarkDictItemsToPython(kwargs)
		if err != nil {
			return starlark.None, err
		}
		defer C.Py_DecRef(ckwargs)

		res := C.PyObject_Call(obj, cargs, ckwargs)
		if C.PyErr_Occurred() != nil {
			return starlark.None, getPyError()
		}

		defer C.Py_DecRef(res)
		return state.innerPythonToStarlarkValue(res)
	}), nil
}

func (state *StarlarkState) pythonToStarlarkMethod(obj *C.PyObject) (starlark.Value, error) {
	self := C.PyMethod_Self(obj)
	f := C.PyMethod_Function(obj)

	funcName, err := getFuncName(f)
	if err != nil {
		return starlark.None, err
	}

	C.Py_IncRef(obj)
	state.childRefs = append(state.childRefs, obj)

	return starlark.NewBuiltin(funcName, func(
		_ *starlark.Thread,
		_ *starlark.Builtin,
		args starlark.Tuple,
		kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		state.ReattachGIL()
		defer state.DetachGIL()

		// create args list with self at the front
		cargsList, err := starlarkTupleToPythonList(args)
		if err != nil {
			return starlark.None, err
		}
		cargs := C.PyTuple_New(C.Py_ssize_t(len(cargsList) + 1))
		if cargs == nil {
			return starlark.None, fmt.Errorf("Could not initialize argument list")
		}
		C.Py_IncRef(self)
		C.PyTuple_SetItem(cargs, 0, self)
		for i, arg := range cargsList {
			C.PyTuple_SetItem(cargs, C.Py_ssize_t(i + 1), arg)
		}
		defer C.Py_DecRef(cargs)

		ckwargs, err := starlarkDictItemsToPython(kwargs)
		if err != nil {
			return starlark.None, err
		}
		defer C.Py_DecRef(ckwargs)

		res := C.PyObject_Call(f, cargs, ckwargs)
		if C.PyErr_Occurred() != nil {
			return starlark.None, getPyError()
		}

		defer C.Py_DecRef(res)
		return state.innerPythonToStarlarkValue(res)
	}), nil
}

func (state *StarlarkState) innerPythonToStarlarkValue(obj *C.PyObject) (starlark.Value, error) {
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
		value, err = state.pythonToStarlarkSet(obj)
	case C.cgoPyDict_Check(obj) == 1:
		value, err = state.pythonToStarlarkDict(obj)
	case C.cgoPyList_Check(obj) == 1:
		value, err = state.pythonToStarlarkList(obj)
	case C.cgoPyTuple_Check(obj) == 1:
		value, err = state.pythonToStarlarkTuple(obj)
	case C.PySequence_Check(obj) == 1:
		value, err = state.pythonToStarlarkList(obj)
	case C.PyMapping_Check(obj) == 1:
		value, err = state.pythonToStarlarkDict(obj)
	case C.cgoPyFunc_Check(obj) == 1:
		value, err = state.pythonToStarlarkFunc(obj)
	case C.cgoPyMethod_Check(obj) == 1:
		value, err = state.pythonToStarlarkMethod(obj)
	default:
		err = fmt.Errorf("Don't know how to convert Python %s to Starlark", C.GoString(obj.ob_type.tp_name))
	}

	if err == nil {
		if C.PyErr_Occurred() != nil {
			err = fmt.Errorf("Python exception while converting to Starlark")
		}
	}

	return value, err
}

func (state *StarlarkState) pythonToStarlarkValue(obj *C.PyObject) (starlark.Value, error) {
	value, err := state.innerPythonToStarlarkValue(obj)
	if err != nil {
		handleConversionError(err, C.ConversionToStarlarkFailed)
		return starlark.None, err
	}

	return value, nil
}
