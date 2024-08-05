#ifndef PYTHON_STARLARK_GO_H
#define PYTHON_STARLARK_GO_H

#include <stdint.h>
#include <stdlib.h>
#define PY_SSIZE_T_CLEAN
#undef Py_LIMITED_API
#include <Python.h>

/* Starlark object */
typedef struct Starlark {
  PyObject_HEAD uintptr_t handle;
} Starlark;

/* Helpers for Cgo, which can't handle varargs or macros */
Starlark *starlarkAlloc(PyTypeObject *type);

void starlarkFree(Starlark *self);

int parseInitArgs(
    PyObject *args, PyObject *kwargs, PyObject **globals, PyObject **print
);

int parseEvalArgs(
    PyObject *args,
    PyObject *kwargs,
    char **expr,
    char **filename,
    unsigned int *convert,
    PyObject **print
);

int parseExecArgs(
    PyObject *args, PyObject *kwargs, char **defs, char **filename, PyObject **print
);

int parseGetGlobalArgs(
    PyObject *args, PyObject *kwargs, char **name, PyObject **default_value
);

int parsePopGlobalArgs(
    PyObject *args, PyObject *kwargs, char **name, PyObject **default_value
);

PyObject *makeStarlarkErrorArgs(const char *error_msg, const char *error_type);

PyObject *makeSyntaxErrorArgs(
    const char *error_msg,
    const char *error_type,
    const char *msg,
    const char *filename,
    const unsigned int line,
    const unsigned int column
);

PyObject *makeEvalErrorArgs(
    const char *error_msg,
    const char *error_type,
    const char *filename,
    const unsigned int line,
    const unsigned int column,
    const char *function_name,
    const char *backtrace
);

PyObject *makeResolveErrorItem(
    const char *msg, const unsigned int line, const unsigned int column
);

PyObject *makeResolveErrorArgs(
    const char *error_msg, const char *error_type, PyObject *errors
);

PyObject *cgoPy_BuildString(const char *src);

PyObject *cgoPy_NewRef(PyObject *obj);

int cgoPyFloat_Check(PyObject *obj);

int cgoPyLong_Check(PyObject *obj);

int cgoPyUnicode_Check(PyObject *obj);

int cgoPyBytes_Check(PyObject *obj);

int cgoPySet_Check(PyObject *obj);

int cgoPyTuple_Check(PyObject *obj);

int cgoPyMapping_Check(PyObject *obj);

int cgoPyDict_Check(PyObject *obj);

int cgoPyList_Check(PyObject *obj);

int cgoPyFunc_Check(PyObject *obj);

int cgoPyMethod_Check(PyObject *obj);

#endif /* PYTHON_STARLARK_GO_H */
