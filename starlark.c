#include "starlark.h"

/* This stuff is in the Go file */
StarlarkGo *StarlarkGo_new(PyTypeObject *type, PyObject *args, PyObject *kwds);
static void StarlarkGo_dealloc(StarlarkGo *self);
PyObject *StarlarkGo_eval(StarlarkGo *self, PyObject *args);
PyObject *StarlarkGo_exec(StarlarkGo *self, PyObject *args);

/* Helpers for Cgo, which can't handle varargs or macros */
StarlarkGo *CgoStarlarkGoAlloc(PyTypeObject *type) {
  return (StarlarkGo *)type->tp_alloc(type, 0);
}

void CgoStarlarkGoDealloc(StarlarkGo *self) {
  Py_TYPE(self)->tp_free((PyObject *)self);
}

PyObject *CgoStarlarkErrorArgs(const char *error_msg, const char *error_type) {
  return Py_BuildValue("ss", error_msg, error_type);
}

PyObject *CgoSyntaxErrorArgs(const char *error_msg, const char *error_type,
                             const char *msg, const char *filename,
                             const unsigned int line,
                             const unsigned int column) {
  return Py_BuildValue("ssssII", error_msg, error_type, msg, filename, line,
                       column);
}

PyObject *CgoEvalErrorArgs(const char *error_msg, const char *error_type,
                           const char *backtrace) {
  return Py_BuildValue("sss", error_msg, error_type, backtrace);
}

void CgoPyDecRef(PyObject *obj) { Py_XDECREF(obj); }

PyObject *CgoPyString(const char *s) { return Py_BuildValue("U", s); }

PyObject *CgoPyNone() { Py_RETURN_NONE; }

PyTypeObject *CgoPyType(PyObject *obj) { return Py_TYPE(obj); }

PyObject *CgoParseEvalArgs(PyObject *args) {
  PyObject *obj;

  if (PyArg_ParseTuple(args, "U", &obj) == 0)
    return NULL;

  return PyUnicode_AsUTF8String(obj);
}

static PyMethodDef StarlarkGo_methods[] = {
    {"eval", (PyCFunction)StarlarkGo_eval, METH_VARARGS,
     "Evaluate a Starlark expression"},
    {"exec", (PyCFunction)StarlarkGo_exec, METH_VARARGS,
     "Execute Starlark code, modifying the global state"},
    {NULL} /* Sentinel */
};

/* Starlark object type */
static PyTypeObject StarlarkGoType = {
    PyVarObject_HEAD_INIT(NULL, 0) // this confuses clang-format
        .tp_name = "pystarlark._lib.StarlarkGo",
    .tp_doc = "Starlark interpreter",
    .tp_basicsize = sizeof(StarlarkGo),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = (newfunc)StarlarkGo_new,
    .tp_dealloc = (destructor)StarlarkGo_dealloc,
    .tp_methods = StarlarkGo_methods};

/* Module */
static PyModuleDef pystarlark_lib = {
    PyModuleDef_HEAD_INIT,
    .m_name = "pystarlark._lib",
    .m_doc = "Interface to starlark-go",
    .m_size = -1,
};

/* Helper to fetch exception classes */
static PyObject *get_exception_class(PyObject *errors, const char *name) {
  PyObject *retval = PyObject_GetAttrString(errors, name);

  if (retval == NULL)
    PyErr_Format(PyExc_RuntimeError, "pystarlark.errors.%s is not defined",
                 name);

  return retval;
}

/* Module initialization */
PyMODINIT_FUNC PyInit__lib(void) {
  PyObject *errors = PyImport_ImportModule("pystarlark.errors");
  if (errors == NULL)
    return NULL;

  StarlarkError = get_exception_class(errors, "StarlarkError");
  if (StarlarkError == NULL)
    return NULL;

  SyntaxError = get_exception_class(errors, "SyntaxError");
  if (SyntaxError == NULL)
    return NULL;

  EvalError = get_exception_class(errors, "EvalError");
  if (EvalError == NULL)
    return NULL;

  PyObject *m;
  if (PyType_Ready(&StarlarkGoType) < 0)
    return NULL;

  m = PyModule_Create(&pystarlark_lib);
  if (m == NULL)
    return NULL;

  Py_INCREF(&StarlarkGoType);
  if (PyModule_AddObject(m, "StarlarkGo", (PyObject *)&StarlarkGoType) < 0) {
    Py_DECREF(&StarlarkGoType);
    Py_DECREF(m);

    return NULL;
  }

  return m;
}
