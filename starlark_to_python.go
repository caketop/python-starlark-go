package main

/*
#include "starlark.h"

extern PyObject *ConversionToPythonFailed;
*/
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"

	"go.starlark.net/starlark"
)

func starlarkIntToPython(x starlark.Int) (*C.PyObject, error) {
	/* Try to do it quickly */
	xint, ok := x.Int64()
	if ok {
		return C.PyLong_FromLongLong(C.longlong(xint)), nil
	}

	/* Fall back to converting from string */
	cstr := C.CString(x.String())
	defer C.free(unsafe.Pointer(cstr))
	return C.PyLong_FromString(cstr, nil, 10), nil
}

func starlarkStringToPython(x starlark.String) (*C.PyObject, error) {
	cstr := C.CString(string(x))
	defer C.free(unsafe.Pointer(cstr))
	return C.cgoPy_BuildString(cstr), nil
}

func starlarkDictToPython(x starlark.IterableMapping) (*C.PyObject, error) {
	items := x.Items()
	return starlarkDictItemsToPython(items)
}

func starlarkDictItemsToPython(items []starlark.Tuple) (*C.PyObject, error) {
	dict := C.PyDict_New()

	for _, item := range items {
		key, err := innerStarlarkValueToPython(item[0])
		if key != nil {
			defer C.Py_DecRef(key)
		}

		if err != nil {
			C.Py_DecRef(dict)
			return nil, fmt.Errorf("While converting key %v in Starlark dict: %v", item[0], err)
		}

		value, err := innerStarlarkValueToPython((item[1]))
		if value != nil {
			defer C.Py_DecRef(value)
		}

		if err != nil {
			C.Py_DecRef(dict)
			return nil, fmt.Errorf("While converting value %v of key %v in Starlark dict: %v", item[1], item[0], err)
		}

		// This does not steal references
		C.PyDict_SetItem(dict, key, value)
	}

	return dict, nil
}

func starlarkTupleToPython(x starlark.Tuple) (*C.PyObject, error) {
	tuple := C.PyTuple_New(C.Py_ssize_t(x.Len()))
	iter := x.Iterate()
	defer iter.Done()

	var elem starlark.Value
	for i := 0; iter.Next(&elem); i++ {
		value, err := innerStarlarkValueToPython(elem)
		if err != nil {
			if value != nil {
				C.Py_DecRef(value)
			}
			C.Py_DecRef(tuple)
			return nil, fmt.Errorf("While converting value %v at index %v in Starlark tuple: %v", elem, i, err)
		}

		// This "steals" the ref to value so we don't need to DecRef after
		if C.PyTuple_SetItem(tuple, C.Py_ssize_t(i), value) != 0 {
			C.Py_DecRef(value)
			C.Py_DecRef(tuple)
			return nil, fmt.Errorf("Couldn't store converted value of %v at index %v in Python tuple: %v", elem, i, err)
		}
	}

	return tuple, nil
}

func starlarkListToPython(x starlark.Iterable) (*C.PyObject, error) {
	list := C.PyList_New(0)
	iter := x.Iterate()
	defer iter.Done()

	var elem starlark.Value
	for i := 0; iter.Next(&elem); i++ {
		value, err := innerStarlarkValueToPython(elem)
		if err != nil {
			C.Py_DecRef(list)
			return nil, fmt.Errorf("While converting value %v at index %v in Starlark list: %v", elem, i, err)
		}

		// This "steals" the ref to value so we don't need to DecRef after
		if C.PyList_Append(list, value) != 0 {
			C.Py_DecRef(value)
			C.Py_DecRef(list)
			return nil, fmt.Errorf("Couldn't store converted value of %v at index %v in Python list: %v", elem, i, err)
		}
	}

	return list, nil
}

func starlarkSetToPython(x starlark.Set) (*C.PyObject, error) {
	set := C.PySet_New(nil)
	iter := x.Iterate()
	defer iter.Done()

	var elem starlark.Value
	for i := 0; iter.Next(&elem); i++ {
		value, err := innerStarlarkValueToPython(elem)
		if value != nil {
			defer C.Py_DecRef(value)
		}

		if err != nil {
			C.Py_DecRef(set)
			return nil, fmt.Errorf("While converting value %v in Starlark set: %v", elem, err)
		}

		// This does not steal references
		C.PySet_Add(set, value)
	}

	return set, nil
}

func starlarkBytesToPython(x starlark.Bytes) (*C.PyObject, error) {
	cstr := C.CString(string(x))
	defer C.free(unsafe.Pointer(cstr))
	return C.PyBytes_FromStringAndSize(cstr, C.Py_ssize_t(x.Len())), nil
}

func innerStarlarkValueToPython(x starlark.Value) (*C.PyObject, error) {
	var value *C.PyObject = nil
	var err error = nil

	switch x := x.(type) {
	case starlark.NoneType:
		value = C.cgoPy_NewRef(C.Py_None)
	case starlark.Bool:
		if x {
			value = C.cgoPy_NewRef(C.Py_True)
		} else {
			value = C.cgoPy_NewRef(C.Py_False)
		}
	case starlark.Int:
		value, err = starlarkIntToPython(x)
	case starlark.Float:
		value = C.PyFloat_FromDouble(C.double(float64(x)))
	case starlark.String:
		value, err = starlarkStringToPython(x)
	case starlark.Bytes:
		value, err = starlarkBytesToPython(x)
	case *starlark.Set:
		value, err = starlarkSetToPython(*x)
	case starlark.IterableMapping:
		value, err = starlarkDictToPython(x)
	case starlark.Tuple:
		value, err = starlarkTupleToPython(x)
	case starlark.Iterable:
		value, err = starlarkListToPython(x)
	default:
		err = fmt.Errorf("Don't know how to convert Starlark %s to Python", reflect.TypeOf(x).String())
	}

	if err == nil {
		if C.PyErr_Occurred() != nil {
			err = fmt.Errorf("Python exception while converting from Starlark")
		}
	}

	return value, err
}

func starlarkValueToPython(x starlark.Value) (*C.PyObject, error) {
	value, err := innerStarlarkValueToPython(x)
	if err != nil {
		handleConversionError(err, C.ConversionToPythonFailed)
		return nil, err
	}

	return value, nil
}
