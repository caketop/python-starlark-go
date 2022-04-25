#include "starlark.h"

/* Declarations for object methods written in Go */
void ConfigureStarlark(int allowSet, int allowGlobalReassign, int allowRecursion);

int Starlark_init(Starlark *self, PyObject *args, PyObject *kwds);
Starlark *Starlark_new(PyTypeObject *type, PyObject *args, PyObject *kwds);
void Starlark_dealloc(Starlark *self);
PyObject *Starlark_eval(Starlark *self, PyObject *args);
PyObject *Starlark_exec(Starlark *self, PyObject *args);
PyObject *Starlark_global_names(Starlark *self, PyObject *_);
PyObject *Starlark_get_global(Starlark *self, PyObject *args, PyObject **kwargs);
PyObject *Starlark_set_globals(Starlark *self, PyObject *args, PyObject **kwargs);
PyObject *Starlark_pop_global(Starlark *self, PyObject *args, PyObject **kwargs);
PyObject *Starlark_get_print(Starlark *self, void *closure);
int Starlark_set_print(Starlark *self, PyObject *value, void *closure);
PyObject *Starlark_tp_iter(Starlark *self);

/* Exceptions - the module init function will fill these in */
PyObject *StarlarkError;
PyObject *SyntaxError;
PyObject *EvalError;
PyObject *ResolveError;
PyObject *ResolveErrorItem;
PyObject *ConversionError;

/* Wrapper for setting Starlark configuration options */
static char *configure_keywords[] = {
    "allow_set", "allow_global_reassign", "allow_recursion", NULL /* Sentinel */
};

PyObject *configure_starlark(PyObject *self, PyObject *args, PyObject *kwargs)
{
  /* ConfigureStarlark interprets -1 as "unspecified" */
  int allow_set = -1, allow_global_reassign = -1, allow_recursion = -1;

  if (PyArg_ParseTupleAndKeywords(
          args,
          kwargs,
          "|$ppp:configure_starlark",
          configure_keywords,
          &allow_set,
          &allow_global_reassign,
          &allow_recursion
      ) == 0) {
    return NULL;
  }

  ConfigureStarlark(allow_set, allow_global_reassign, allow_recursion);
  Py_RETURN_NONE;
}

/* Container for module methods */
static PyMethodDef module_methods[] = {
    {"configure_starlark",
     (PyCFunction)configure_starlark,
     METH_VARARGS | METH_KEYWORDS,
     "Configure the starlark interpreter"},
    {NULL} /* Sentinel */
};

/* Container for object methods */
static PyMethodDef StarlarkGo_methods[] = {
    {"eval",
     (PyCFunction)Starlark_eval,
     METH_VARARGS | METH_KEYWORDS,
     "Evaluate a Starlark expression"},
    {"exec",
     (PyCFunction)Starlark_exec,
     METH_VARARGS | METH_KEYWORDS,
     "Execute Starlark code, modifying the global state"},
    {"globals", (PyCFunction)Starlark_global_names, METH_NOARGS, "TODO"},
    {"get", (PyCFunction)Starlark_get_global, METH_VARARGS | METH_KEYWORDS, "TODO"},
    {"set", (PyCFunction)Starlark_set_globals, METH_VARARGS | METH_KEYWORDS, "TODO"},
    {"pop", (PyCFunction)Starlark_pop_global, METH_VARARGS | METH_KEYWORDS, "TODO"},
    {NULL} /* Sentinel */
};

static PyGetSetDef Starlark_getset[] = {
    {"print", (getter)Starlark_get_print, (setter)Starlark_set_print, "TODO", NULL},
    {NULL},
};

/* Python type for object */
static PyTypeObject StarlarkType = {
    // clang-format off
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "pystarlark.starlark_go.Starlark",
    // clang-format on
    .tp_doc = "Starlark interpreter",
    .tp_basicsize = sizeof(Starlark),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = (newfunc)Starlark_new,
    .tp_init = (initproc)Starlark_init,
    .tp_dealloc = (destructor)Starlark_dealloc,
    .tp_methods = StarlarkGo_methods,
    .tp_iter = (getiterfunc)Starlark_tp_iter,
    .tp_getset = Starlark_getset,
};

/* Module */
static PyModuleDef pystarlark_lib = {
    PyModuleDef_HEAD_INIT,
    .m_name = "pystarlark.starlark_go",
    .m_doc = "Interface to starlark-go",
    .m_size = -1,
    .m_methods = module_methods,
};

/* Argument names for our methods */
static char *init_keywords[] = {"globals", "print", NULL};
static char *eval_keywords[] = {"expr", "filename", "convert", "print", NULL};
static char *exec_keywords[] = {"defs", "filename", "print", NULL};
static char *get_global_keywords[] = {"name", "default", NULL};

/* Helpers to allocate and free our object */
Starlark *starlarkAlloc(PyTypeObject *type)
{
  /* Necessary because Cgo can't do function pointers */
  return (Starlark *)type->tp_alloc(type, 0);
}

void starlarkFree(Starlark *self)
{
  /* Necessary because Cgo can't do function pointers */
  Py_TYPE(self)->tp_free((PyObject *)self);
}

/* Helpers to parse method arguments */
int parseInitArgs(
    PyObject *args, PyObject *kwargs, PyObject **globals, PyObject **print
)
{
  /* Necessary because Cgo can't do varargs */
  /* One optional object */
  return PyArg_ParseTupleAndKeywords(
      args, kwargs, "|$OO:Starlark", init_keywords, globals, print
  );
}

int parseEvalArgs(
    PyObject *args,
    PyObject *kwargs,
    char **expr,
    char **filename,
    unsigned int *convert,
    PyObject **print
)
{
  /* Necessary because Cgo can't do varargs */
  /* One required string, folloed by an optional string and an optional bool */
  return PyArg_ParseTupleAndKeywords(
      args, kwargs, "s|$spO:eval", eval_keywords, expr, filename, convert, print
  );
}

int parseExecArgs(
    PyObject *args, PyObject *kwargs, char **defs, char **filename, PyObject **print
)
{
  /* Necessary because Cgo can't do varargs */
  /* One required string, folloed by an optional string */
  return PyArg_ParseTupleAndKeywords(
      args, kwargs, "s|$sO:exec", exec_keywords, defs, filename, print
  );
}

int parseGetGlobalArgs(
    PyObject *args, PyObject *kwargs, char **name, PyObject **default_value
)
{
  /* Necessary because Cgo can't do varargs */
  /* One required string, full stop */
  return PyArg_ParseTupleAndKeywords(
      args, kwargs, "s|O:get", get_global_keywords, name, default_value
  );
}

int parsePopGlobalArgs(
    PyObject *args, PyObject *kwargs, char **name, PyObject **default_value
)
{
  /* Necessary because Cgo can't do varargs */
  /* One required string, full stop */
  return PyArg_ParseTupleAndKeywords(
      args, kwargs, "s|O:pop", get_global_keywords, name, default_value
  );
}

/* Helpers for Cgo to build exception arguments */
PyObject *makeStarlarkErrorArgs(const char *error_msg, const char *error_type)
{
  /* Necessary because Cgo can't do varargs */
  return Py_BuildValue("ss", error_msg, error_type);
}

PyObject *makeSyntaxErrorArgs(
    const char *error_msg,
    const char *error_type,
    const char *msg,
    const char *filename,
    const unsigned int line,
    const unsigned int column
)
{
  /* Necessary because Cgo can't do varargs */
  /* Four strings and two unsigned integers */
  return Py_BuildValue("ssssII", error_msg, error_type, msg, filename, line, column);
}

PyObject *makeEvalErrorArgs(
    const char *error_msg, const char *error_type, const char *backtrace
)
{
  /* Necessary because Cgo can't do varargs */
  /* Three strings */
  return Py_BuildValue("sss", error_msg, error_type, backtrace);
}

PyObject *makeResolveErrorItem(
    const char *msg, const unsigned int line, const unsigned int column
)
{
  /* Necessary because Cgo can't do varargs */
  /* A string and two unsigned integers */
  PyObject *args = Py_BuildValue("sII", msg, line, column);
  PyObject *obj = PyObject_CallObject(ResolveErrorItem, args);
  Py_DECREF(args);
  return obj;
}

PyObject *makeResolveErrorArgs(
    const char *error_msg, const char *error_type, PyObject *errors
)
{
  /* Necessary because Cgo can't do varargs */
  /* Two strings and a Python object */
  return Py_BuildValue("ssO", error_msg, error_type, errors);
}

/* Other assorted helpers for Cgo */
PyObject *cgoPy_BuildString(const char *src)
{
  /* Necessary because Cgo can't do varargs */
  return Py_BuildValue("s", src);
}

PyObject *cgoPy_NewRef(PyObject *obj)
{
  /* Necessary because Cgo can't do macros and Py_NewRef is part of
   * Python's "stable API" but only since 3.10
   */
  Py_INCREF(obj);
  return obj;
}

int cgoPyFloat_Check(PyObject *obj)
{
  /* Necessary because Cgo can't do macros */
  return PyFloat_Check(obj);
}

int cgoPyLong_Check(PyObject *obj)
{
  /* Necessary because Cgo can't do macros */
  return PyLong_Check(obj);
}

int cgoPyUnicode_Check(PyObject *obj)
{
  /* Necessary because Cgo can't do macros */
  return PyUnicode_Check(obj);
}

int cgoPyBytes_Check(PyObject *obj)
{
  /* Necessary because Cgo can't do macros */
  return PyBytes_Check(obj);
}

int cgoPySet_Check(PyObject *obj)
{
  /* Necessary because Cgo can't do macros */
  return PySet_Check(obj);
}

int cgoPyTuple_Check(PyObject *obj)
{
  /* Necessary because Cgo can't do macros */
  return PyTuple_Check(obj);
}

int cgoPyDict_Check(PyObject *obj)
{
  /* Necessary because Cgo can't do macros */
  return PyDict_Check(obj);
}

int cgoPyList_Check(PyObject *obj)
{
  /* Necessary because Cgo can't do macros */
  return PyList_Check(obj);
}

/* Helper to fetch exception classes */
static PyObject *get_exception_class(PyObject *errors, const char *name)
{
  PyObject *retval = PyObject_GetAttrString(errors, name);

  if (retval == NULL)
    PyErr_Format(PyExc_RuntimeError, "pystarlark.errors.%s is not defined", name);

  return retval;
}

/* Module initialization */
PyMODINIT_FUNC PyInit_starlark_go(void)
{
  PyObject *errors = PyImport_ImportModule("pystarlark.errors");
  if (errors == NULL) return NULL;

  StarlarkError = get_exception_class(errors, "StarlarkError");
  if (StarlarkError == NULL) return NULL;

  SyntaxError = get_exception_class(errors, "SyntaxError");
  if (SyntaxError == NULL) return NULL;

  EvalError = get_exception_class(errors, "EvalError");
  if (EvalError == NULL) return NULL;

  ResolveError = get_exception_class(errors, "ResolveError");
  if (ResolveError == NULL) return NULL;

  ResolveErrorItem = get_exception_class(errors, "ResolveErrorItem");
  if (ResolveErrorItem == NULL) return NULL;

  ConversionError = get_exception_class(errors, "ConversionError");
  if (ConversionError == NULL) return NULL;

  PyObject *m;
  if (PyType_Ready(&StarlarkType) < 0) return NULL;

  m = PyModule_Create(&pystarlark_lib);
  if (m == NULL) return NULL;

  Py_INCREF(&StarlarkType);
  if (PyModule_AddObject(m, "Starlark", (PyObject *)&StarlarkType) < 0) {
    Py_DECREF(&StarlarkType);
    Py_DECREF(m);

    return NULL;
  }

  return m;
}
