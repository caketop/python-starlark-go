#ifndef PYTHON_STARLARK_GO_H
#define PYTHON_STARLARK_GO_H

#define PY_SSIZE_T_CLEAN
#include <Python.h>

/* Starlark object */
typedef struct StarlarkGo {
  PyObject_HEAD unsigned long starlark_thread;
} StarlarkGo;

/* Custom exceptions */
PyObject *StarlarkError;
PyObject *SyntaxError;
PyObject *EvalError;

/* Helpers for Cgo, which can't handle varargs or macros */
StarlarkGo *CgoStarlarkGoAlloc(PyTypeObject *type);

void CgoStarlarkGoDealloc(StarlarkGo *self);

PyObject *CgoStarlarkErrorArgs(const char *error_msg, const char *error_type);

PyObject *CgoSyntaxErrorArgs(const char *error_msg, const char *error_type,
                             const char *msg, const char *filename,
                             const unsigned int line,
                             const unsigned int column);

PyObject *CgoEvalErrorArgs(const char *error_msg, const char *error_type,
                           const char *backtrace);

void CgoPyDecRef(PyObject *obj);

PyObject *CgoPyString(const char *s);

PyObject *CgoPyNone();

PyObject *CgoParseEvalArgs(PyObject *args);

PyTypeObject *CgoPyType(PyObject *obj);

#endif /* PYTHON_STARLARK_GO_H */
