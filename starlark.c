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
PyObject *ConversionToPythonFailed;
PyObject *ConversionToStarlarkFailed;

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

PyDoc_STRVAR(
    configure_starlark_doc,
    "configure_starlark(*, allow_set=None, allow_global_reassign=None, "
    "allow_recursion=None)\n--\n\n"
    "Change what features the Starlark interpreter allows. Unfortunately, "
    "this manipulates global variables, and affects all Starlark interpreters "
    "in your application. It is not possible to have one Starlark "
    "interpreter with ``allow_set=True`` and another with ``allow_set=False`` "
    "simultaneously.\n\n"
    "All feature flags are initially ``False``.\n\n"
    "See the `starlark-go documentation "
    "<https://pkg.go.dev/go.starlark.net/resolve#pkg-variables>`_ for "
    "more information.\n\n"
    ":param allow_set: If ``True``, allow the creation of `set "
    "<https://github.com/google/starlark-go/blob/master/doc/spec.md#sets=>`_ objects "
    "in Starlark.\n"
    ":type allow_set: typing.Optional[bool]\n"
    ":param allow_global_reassign: If ``True``, allow reassignment to top-level names; "
    "also, allow if/for/while at top-level.\n"
    ":type allow_global_reassign:  typing.Optional[bool]\n"
    ":param allow_recursion: If ``True``, allow while statements and recursive "
    "functions.\n"
    ":type allow_recursion:  typing.Optional[bool]\n"
);

/* Argument names and documentation for our methods */
static char *init_keywords[] = {"globals", "print", NULL};

PyDoc_STRVAR(
    Starlark_init_doc,
    "Starlark(*, globals=None, print=None)\n--\n\n"
    "Create a Starlark object. A Starlark object contains a set of global variables, "
    "which can be manipulated by executing Starlark code.\n\n"
    ":param globals: Initial set of global variables. Keys must be strings. Values can "
    "be any type supported by :func:`set`.\n"
    ":type globals: typing.Mapping[str, typing.Any]\n"
    ":param print: A function to call in place of Starlark's ``print()`` function. If "
    "unspecified, Starlark's ``print()`` function will be forwarded to Python's "
    "built-in :py:func:`python:print`.\n"
    ":type print: typing.Callable[[str], typing.Any]\n"
);

static char *eval_keywords[] = {"expr", "filename", "convert", "print", NULL};

PyDoc_STRVAR(
    Starlark_eval_doc,
    "eval(self, expr, *, filename=None, convert=True, print=None)\n--\n\n"
    "Evaluate a Starlark expression. The expression passed to ``eval`` must evaluate "
    "to a value. Function definitions, variable assignments, and control structures "
    "are not allowed by ``eval``. To use those, please use :func:`.exec`.\n\n"
    ":param expr: A string containing a Starlark expression to evaluate\n"
    ":type expr: str\n"
    ":param filename: An optional filename to use in exceptions, if evaluting the "
    "expression fails.\n"
    ":type filename: typing.Optional[str]\n"
    ":param convert: If True, convert the result of the expression into a Python "
    "value. If False, return a string containing the representation of the expression "
    "in Starlark. Defaults to True.\n"
    ":type convert: bool\n"
    ":param print: A function to call in place of Starlark's ``print()`` function. If "
    "unspecified, Starlark's ``print()`` function will be forwarded to Python's "
    "built-in :py:func:`python:print`.\n"
    ":type print: typing.Callable[[str], typing.Any]\n"
    ":raises ConversionToPythonFailed: if the value is of an unsupported type for "
    "conversion.\n"
    ":raises EvalError: if there is a Starlark evaluation error\n"
    ":raises ResolveError: if there is a Starlark resolution error\n"
    ":raises SyntaxError: if there is a Starlark syntax error\n"
    ":raises StarlarkError: if there is an unexpected error\n"
    ":rtype: typing.Any\n"
);

static char *exec_keywords[] = {"defs", "filename", "print", NULL};

PyDoc_STRVAR(
    Starlark_exec_doc,
    "exec(self, defs, *, filename=None, print=None)\n--\n\n"
    "Execute Starlark code. All legal Starlark constructs may be used with "
    "``exec``.\n\n"
    "``exec`` does not return a value. To evaluate the value of a Starlark expression, "
    "please use func:`eval`.\n\n"
    ":param defs: A string containing Starlark code to execute\n"
    ":type defs: str\n"
    ":param filename: An optional filename to use in exceptions, if evaluting the "
    "expression fails.\n"
    ":type filename: Optional[str]\n"
    ":param print: A function to call in place of Starlark's ``print()`` function. If "
    "unspecified, Starlark's ``print()`` function will be forwarded to Python's "
    "built-in :py:func:`python:print`.\n"
    ":type print: typing.Callable[[str], typing.Any]\n"
    ":raises EvalError: if there is a Starlark evaluation error\n"
    ":raises ResolveError: if there is a Starlark resolution error\n"
    ":raises SyntaxError: if there is a Starlark syntax error\n"
    ":raises StarlarkError: if there is an unexpected error\n"
);

static char *get_global_keywords[] = {"name", "default", NULL};

PyDoc_STRVAR(
    Starlark_get_doc,
    "get(self, name, default_value = ...)\n--\n\n"
    "Get the value of a Starlark global variable.\n\n"
    "Conversion from most Starlark data types is supported:\n\n"
    "* Starlark `None <https://pkg.go.dev/go.starlark.net/starlark#None>`_ to "
    "Python :py:obj:`python:None`\n"
    "* Starlark `bool <https://pkg.go.dev/go.starlark.net/starlark#Bool>`_ to "
    "Python :py:obj:`python:bool`\n"
    "* Starlark `bytes <https://pkg.go.dev/go.starlark.net/starlark#Bytes>`_ to "
    "Python :py:obj:`python:bytes`\n"
    "* Starlark `float <https://pkg.go.dev/go.starlark.net/starlark#Float>`_ to "
    "Python :py:obj:`python:float`\n"
    "* Starlark `int <https://pkg.go.dev/go.starlark.net/starlark#Int>`_ to "
    "Python :py:obj:`python:int`\n"
    "* Starlark `string <https://pkg.go.dev/go.starlark.net/starlark#String>`_ to "
    "Python :py:obj:`python:str`\n"
    "* Starlark `dict <https://pkg.go.dev/go.starlark.net/starlark#Dict>`_ (and "
    "`IterableMapping <https://pkg.go.dev/go.starlark.net/starlark#IterableMapping>`_) "
    "to Python :py:obj:`python:dict`\n"
    "* Starlark `list <https://pkg.go.dev/go.starlark.net/starlark#List>`_ (and "
    "`Iterable <https://pkg.go.dev/go.starlark.net/starlark#Iterable>`_) to "
    "Python :py:obj:`python:list`\n"
    "* Starlark `set <https://pkg.go.dev/go.starlark.net/starlark#Set>`_ to "
    "Python :py:obj:`python:set`\n"
    "* Starlark `tuple <https://pkg.go.dev/go.starlark.net/starlark#Tuple>`_ to "
    "Python :py:obj:`python:tuple`\n\n"
    "For the aggregate types (``dict``, ``list``, ``set``, and ``tuple``,) all keys "
    "and/or values must also be one of the supported types.\n\n"
    "Attempting to get the value of any other Starlark type will raise a "
    ":py:class:`ConversionToPythonFailed`.\n\n"
    ":param name: The name of the global variable.\n"
    ":type name: str\n"
    ":param default_value: A default value to return, if no global variable named "
    "``name`` is defined.\n"
    ":type default_value: typing.Any\n"
    ":raises KeyError: if there is no global value named ``name`` defined.\n"
    ":raises ConversionToPythonFailed: if the value is of an unsupported type for "
    "conversion.\n"
    ":rtype: typing.Any\n"
);

PyDoc_STRVAR(
    Starlark_globals_doc,
    "globals(self)\n--\n\n"
    "Get the names of the currently defined global variables.\n\n"
    ":rtype: typing.List[str]\n"
);

PyDoc_STRVAR(
    Starlark_set_doc,
    "set(self, **kwargs)\n--\n\n"
    "Set the value of one or more Starlark global variables.\n\n"
    "For each keyword parameter specified, one global variable is set.\n\n"
    "Conversion from most basic Python types is supported:\n\n"
    "* Python :py:obj:`python:None` to Starlark `None "
    "<https://pkg.go.dev/go.starlark.net/starlark#None>`_\n"
    "* Python :py:obj:`python:bool` to Starlark `bool "
    "<https://pkg.go.dev/go.starlark.net/starlark#Bool>`_\n"
    "* Python :py:obj:`python:bytes` to Starlark `bytes "
    "<https://pkg.go.dev/go.starlark.net/starlark#Bytes>`_\n"
    "* Python :py:obj:`python:float` to Starlark `float "
    "<https://pkg.go.dev/go.starlark.net/starlark#Float>`_\n"
    "* Python :py:obj:`python:int` to Starlark `int "
    "<https://pkg.go.dev/go.starlark.net/starlark#Int>`_\n"
    "* Python :py:obj:`python:str` to Starlark `string "
    "<https://pkg.go.dev/go.starlark.net/starlark#String>`_\n"
    "* Python :py:obj:`python:dict` (and other objects that implement the mapping "
    "protocol) to Starlark "
    "`dict <https://pkg.go.dev/go.starlark.net/starlark#Dict>`_\n"
    "* Python :py:obj:`python:list` (and other objects that implement the sequence "
    "protocol) to Starlark "
    "`list <https://pkg.go.dev/go.starlark.net/starlark#List>`_\n"
    "* Python :py:obj:`python:set` to Starlark `set "
    "<https://pkg.go.dev/go.starlark.net/starlark#Set>`_\n"
    "* Python :py:obj:`python:tuple` to Starlark `tuple "
    "<https://pkg.go.dev/go.starlark.net/starlark#Tuple>`_\n\n"
    "For the aggregate types (``dict``, ``list``, ``set``, and ``tuple``,) all keys "
    "and/or values must also be one of the supported types.\n\n"
    "Attempting to set a value of any other Python type will raise a "
    ":py:class:`ConversionToStarlarkFailed`.\n\n"
    ":raises ConversionToStarlarkFailed: if a value is of an unsupported type for "
    "conversion.\n"
);

PyDoc_STRVAR(
    Starlark_pop_doc,
    "pop(self, name, default_value = ...)\n--\n\n"
    "Remove a Starlark global variable, and return its value.\n\n"
    "If a value of ``name`` does not exist, and no ``default_value`` has been "
    "specified, raise :py:obj:`python:KeyError`. Otherwise, return "
    "``default_value``.\n\n"
    ":param name: The name of the global variable.\n"
    ":type name: str\n"
    ":param default_value: A default value to return, if no global variable named "
    "``name`` is defined.\n"
    ":type default_value: typing.Any\n"
    ":raises KeyError: if there is no global value named ``name`` defined.\n"
    ":raises ConversionToPythonFailed: if the value is of an unsupported type for "
    "conversion.\n"
    ":rtype: typing.Any\n"
);

PyDoc_STRVAR(
    Starlark_print_doc,
    "A function to call in place of Starlark's ``print()`` function. If "
    "unspecified, Starlark's ``print()`` function will be forwarded to Python's "
    "built-in :py:func:`python:print`.\n\n"
    ":type: typing.Callable[[str], typing.Any]\n"
);

/* Container for module methods */
static PyMethodDef module_methods[] = {
    {"configure_starlark",
     (PyCFunction)configure_starlark,
     METH_VARARGS | METH_KEYWORDS,
     configure_starlark_doc},
    {NULL} /* Sentinel */
};

/* Container for object methods */
static PyMethodDef StarlarkGo_methods[] = {
    {"eval",
     (PyCFunction)Starlark_eval,
     METH_VARARGS | METH_KEYWORDS,
     Starlark_eval_doc},
    {"exec",
     (PyCFunction)Starlark_exec,
     METH_VARARGS | METH_KEYWORDS,
     Starlark_exec_doc},
    {"globals", (PyCFunction)Starlark_global_names, METH_NOARGS, Starlark_globals_doc},
    {"get",
     (PyCFunction)Starlark_get_global,
     METH_VARARGS | METH_KEYWORDS,
     Starlark_get_doc},
    {"set",
     (PyCFunction)Starlark_set_globals,
     METH_VARARGS | METH_KEYWORDS,
     Starlark_set_doc},
    {"pop",
     (PyCFunction)Starlark_pop_global,
     METH_VARARGS | METH_KEYWORDS,
     Starlark_pop_doc},
    {NULL} /* Sentinel */
};

static PyGetSetDef Starlark_getset[] = {
    {"print",
     (getter)Starlark_get_print,
     (setter)Starlark_set_print,
     Starlark_print_doc,
     NULL},
    {NULL},
};

/* Python type for object */
static PyTypeObject StarlarkType = {
    // clang-format off
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "starlark_go.starlark_go.Starlark",
    // clang-format on
    .tp_doc = Starlark_init_doc,
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
static PyModuleDef starlark_go = {
    PyModuleDef_HEAD_INIT,
    .m_name = "starlark_go.starlark_go",
    .m_doc = "Interface to starlark-go",
    .m_size = -1,
    .m_methods = module_methods,
};

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
    const char *error_msg,
    const char *error_type,
    const char *filename,
    const unsigned int line,
    const unsigned int column,
    const char *function_name,
    const char *backtrace
)
{
  /* Necessary because Cgo can't do varargs */
  /* Three strings */
  return Py_BuildValue(
      "sssIIss", error_msg, error_type, filename, line, column, function_name, backtrace
  );
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
    PyErr_Format(PyExc_RuntimeError, "starlark_go.errors.%s is not defined", name);

  return retval;
}

/* Module initialization */
PyMODINIT_FUNC PyInit_starlark_go(void)
{
  PyObject *errors = PyImport_ImportModule("starlark_go.errors");
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

  ConversionToPythonFailed = get_exception_class(errors, "ConversionToPythonFailed");
  if (ConversionToPythonFailed == NULL) return NULL;

  ConversionToStarlarkFailed =
      get_exception_class(errors, "ConversionToStarlarkFailed");
  if (ConversionToStarlarkFailed == NULL) return NULL;

  PyObject *m;
  if (PyType_Ready(&StarlarkType) < 0) return NULL;

  m = PyModule_Create(&starlark_go);
  if (m == NULL) return NULL;

  Py_INCREF(&StarlarkType);
  if (PyModule_AddObject(m, "Starlark", (PyObject *)&StarlarkType) < 0) {
    Py_DECREF(&StarlarkType);
    Py_DECREF(m);

    return NULL;
  }

  return m;
}
