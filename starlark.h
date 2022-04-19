#ifndef PYTHON_STARLARK_GO_H
#define PYTHON_STARLARK_GO_H

#define PY_SSIZE_T_CLEAN
#include <Python.h>

/* Starlark object */
typedef struct StarlarkGo {
  PyObject_HEAD unsigned long starlark_thread;
} StarlarkGo;

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

PyObject *CgoResolveErrorItem(const char *msg, const unsigned int line,
                              const unsigned int column);

PyObject *CgoResolveErrorArgs(const char *error_msg, const char *error_type,
                              PyObject *errors);
PyObject *CgoPyBuildOneValue(const char *fmt, const void *src);

PyObject *CgoPyNone();

PyObject *CgoPyNewRef(PyObject *obj);

int CgoParseEvalArgs(PyObject *args, PyObject *kwargs, char **expr,
                     char **filename, unsigned int *parse);

int GgoParseExecArgs(PyObject *args, PyObject *kwargs, char **defs,
                     char **filename);

#endif /* PYTHON_STARLARK_GO_H */
