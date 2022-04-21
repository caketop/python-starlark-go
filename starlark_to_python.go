package main

/*
#include "starlark.h"

extern PyObject *ConversionError;
*/
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"

	"go.starlark.net/starlark"
)

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
		key := starlarkValueToPython(item[0])
		defer C.Py_DecRef(key)

		if key == nil {
			C.Py_DecRef(dict)
			return nil
		}

		value := starlarkValueToPython((item[1]))
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
		value := starlarkValueToPython(elem)

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
		value := starlarkValueToPython(elem)

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
		value := starlarkValueToPython(elem)
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

func starlarkValueToPython(x starlark.Value) *C.PyObject {
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
