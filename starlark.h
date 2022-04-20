#ifndef PYTHON_STARLARK_GO_H
#define PYTHON_STARLARK_GO_H

#define PY_SSIZE_T_CLEAN
#include <Python.h>

/* Starlark object */
typedef struct Starlark {
  PyObject_HEAD unsigned long state_id;
} Starlark;

/* Helpers for Cgo, which can't handle varargs or macros */
Starlark *starlarkAlloc(PyTypeObject *type);

void starlarkFree(Starlark *self);

int parseEvalArgs(PyObject *args, PyObject *kwargs, char **expr,
                  char **filename, unsigned int *convert);

int parseExecArgs(PyObject *args, PyObject *kwargs, char **defs,
                  char **filename);

PyObject *makeStarlarkErrorArgs(const char *error_msg, const char *error_type);

PyObject *makeSyntaxErrorArgs(const char *error_msg, const char *error_type,
                              const char *msg, const char *filename,
                              const unsigned int line,
                              const unsigned int column);

PyObject *makeEvalErrorArgs(const char *error_msg, const char *error_type,
                            const char *backtrace);

PyObject *makeResolveErrorItem(const char *msg, const unsigned int line,
                               const unsigned int column);

PyObject *makeResolveErrorArgs(const char *error_msg, const char *error_type,
                               PyObject *errors);

PyObject *cgoPy_BuildString(const char *src);

PyObject *cgoPy_NewRef(PyObject *obj);

#endif /* PYTHON_STARLARK_GO_H */
