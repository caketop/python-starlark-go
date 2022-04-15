#define PY_SSIZE_T_CLEAN
#include <Python.h>


/* This stuff is in the Go file */
unsigned long NewThread();
void DestroyThread(unsigned long threadId);
char *Eval(unsigned long threadId, char *stmt, void **pyThread);
int ExecFile(unsigned long threadId, char *data, void **pyThread);
void FreeCString(char *s);

/* Custom exceptions */
static PyObject *StarlarkError = NULL;
static PyObject *SyntaxError = NULL;
static PyObject *EvalError = NULL;

static inline PyObject *PyUnicode_Copy(const char *str) {
    int length = strlen(str);

    PyObject *pystr = PyUnicode_New(length, 1114111);
    PyUnicode_CopyCharacters(pystr, 0, PyUnicode_FromString(str), 0, length);

    return pystr;
}
/* Helpers to raise custom exceptions from Go */
void Raise_StarlarkError(const char *error, const char *error_type) {
    PyThreadState *thread = PyEval_SaveThread();
    PyGILState_STATE gilstate = PyGILState_Ensure();
    PyObject *exc_args = PyTuple_New(2);

	PyTuple_SetItem(exc_args, 0, PyUnicode_Copy(error));
	PyTuple_SetItem(exc_args, 1, PyUnicode_Copy(error_type));
	PyErr_SetObject(StarlarkError, exc_args);
    PyGILState_Release(gilstate);
    PyEval_RestoreThread(thread);
}

void Raise_SyntaxError(const char *error, const char *error_type, const char *msg, const char *filename, const long line, const long column) {
    PyThreadState *thread = PyEval_SaveThread();
    PyGILState_STATE gilstate = PyGILState_Ensure();
    PyObject *exc_args = PyTuple_New(6);

	PyTuple_SetItem(exc_args, 0, PyUnicode_Copy(error));
	PyTuple_SetItem(exc_args, 1, PyUnicode_Copy(error_type));
	PyTuple_SetItem(exc_args, 2, PyUnicode_Copy(msg));
	PyTuple_SetItem(exc_args, 3, PyUnicode_Copy(filename));
	PyTuple_SetItem(exc_args, 4, PyLong_FromLong(line));
	PyTuple_SetItem(exc_args, 5, PyLong_FromLong(column));
	PyErr_SetObject(SyntaxError, exc_args);
    PyGILState_Release(gilstate);
    PyEval_RestoreThread(thread);
}

void Raise_EvalError(const char *error, const char *error_type, const char *backtrace) {
    PyThreadState *thread = PyEval_SaveThread();
    PyGILState_STATE gilstate = PyGILState_Ensure();
    PyObject *exc_args = PyTuple_New(3);

	PyTuple_SetItem(exc_args, 0, PyUnicode_Copy(error));
	PyTuple_SetItem(exc_args, 1, PyUnicode_Copy(error_type));
	PyTuple_SetItem(exc_args, 2, PyUnicode_Copy(backtrace));
	PyErr_SetObject(EvalError, exc_args);
    PyGILState_Release(gilstate);
    PyEval_RestoreThread(thread);
}

/* Helpers to process custom exception arguments */
static inline PyObject *get_arg(PyObject *self, int index) {
    if (!PyObject_HasAttrString(self, "args"))
        return NULL;

	PyObject *args = PyObject_GetAttrString(self, "args");
    PyObject *arg = NULL;

    if (PyTuple_Size(args) > index)
        arg = PyTuple_GetItem(args, index);

    Py_XDECREF(args);
    return arg;
}

static inline PyObject *get_arg_str(PyObject *self, int index) {
    PyObject *arg = get_arg(self, index);
    if (arg == NULL)
        return PyUnicode_FromString("???");

    return arg;
}

static inline PyObject *get_arg_int(PyObject *self, int index) {
    PyObject *arg = get_arg(self, index);
    if ((arg == NULL) || (!PyLong_Check(arg)))
        return PyLong_FromLong(-1);

    return arg;
}

/* Implementation of __str__ for StarlarkError and subclasses */
static PyObject *StarlarkError_str(PyObject *self) {
    return get_arg_str(self, 0);
}

/* StarlarkError properties */
static PyObject *StarlarkError_get_error(PyObject *self, void *closure) {
    return get_arg_str(self, 0);
}

static PyObject *StarlarkError_get_error_type(PyObject *self, void *closure) {
    return get_arg_str(self, 1);
}

static PyGetSetDef StarlarkError_getset[] = {
    {"error", StarlarkError_get_error, NULL, "Summary of the error", NULL},
    {"error_type", StarlarkError_get_error_type, NULL, "Name of the Go type of the error", NULL},
    {NULL}
};

/* SyntaxError properties */
static PyObject *SyntaxError_get_msg(PyObject *self, void *closure) {
    return get_arg_str(self, 2);
}

static PyObject *SyntaxError_get_filename(PyObject *self, void *closure) {
    return get_arg_str(self, 3);
}

static PyObject *SyntaxError_get_line(PyObject *self, void *closure) {
    return get_arg_int(self, 4);
}

static PyObject *SyntaxError_get_column(PyObject *self, void *closure) {
    return get_arg_int(self, 5);
}

static PyGetSetDef SyntaxError_getset[] = {
    {"msg", SyntaxError_get_msg, NULL, "Description of the error", NULL},
    {"filename", SyntaxError_get_filename, NULL, "The name of the file in which the error occurred", NULL},
    {"line", SyntaxError_get_line, NULL, "The line on which the error occurred", NULL},
    {"column", SyntaxError_get_column, NULL, "The column on which the error occurred", NULL},
    {NULL}
};

/* EvalError properties */
static PyObject *EvalError_get_backtrace(PyObject *self, void *closure) {
    return get_arg_str(self, 2);
}

static PyGetSetDef EvalError_getset[] = {
    {"backtrace", EvalError_get_backtrace, NULL, "Trace of execution leading to error", NULL},
    {NULL}
};

/* Helper to add getters to a type */
static int add_getset(PyTypeObject *type, PyGetSetDef *getsets) {
    int retval = 1;

    for (PyGetSetDef *getset = getsets; getset->name; getset++) {
        PyObject *descr = PyDescr_NewGetSet(type, getset);

        if (PyDict_SetItem(type->tp_dict, PyDescr_NAME(descr), descr) < 0) {
            PyErr_SetString(PyExc_RuntimeError, "failed to add getset");
            retval = 0;
        }

        Py_DECREF(descr);
        if (retval != 1)
            break;
    }

    return retval;
}

/* Starlark object */
typedef struct {
    PyObject_HEAD
    unsigned long starlark_thread;
} StarlarkObject;


/* Starlark object methods */
static PyObject* Starlark_new(PyTypeObject *type, PyObject *args, PyObject *kwds) {
    StarlarkObject *self;
    self = (StarlarkObject *) type->tp_alloc(type, 0);

    if (self != NULL) {
        self->starlark_thread = NewThread();
    }

    return (PyObject *) self;
}

static void Starlark_dealloc(StarlarkObject *self) {
    DestroyThread(self->starlark_thread);
    Py_TYPE(self)->tp_free((PyObject *) self);
}


static PyObject* Starlark_eval(StarlarkObject *self, PyObject *args) {
    PyObject *obj;
    PyObject *stmt;
    char *cvalue;
    PyObject *value;
    evalReturn retval;

    if (PyArg_ParseTuple(args, "U", &obj) == 0)
        return NULL;

    stmt = PyUnicode_AsUTF8String(obj);
    if (stmt == NULL)
        return NULL;

    cvalue = Eval(self->starlark_thread, PyBytes_AsString(stmt), &evalReturn);

    if (cvalue == NULL)
    {
        value = NULL;
    }
    else
    {
        value = PyUnicode_FromString(cvalue);
        FreeCString(cvalue);
    }

    Py_DecRef(stmt);

    return value;
}

static PyObject* Starlark_exec(StarlarkObject *self, PyObject *args) {
    PyObject *obj;
    PyObject *data;
    int rc;

    if (PyArg_ParseTuple(args, "U", &obj) == 0)
        return NULL;

    data = PyUnicode_AsUTF8String(obj);
    if (data == NULL)
        return NULL;

    rc = ExecFile(self->starlark_thread, PyBytes_AsString(data));
    Py_DecRef(data);

    if (!rc)
        return NULL;

    Py_RETURN_NONE;
}

static PyMethodDef Starlark_methods[] = {
    {"eval", (PyCFunction) Starlark_eval, METH_VARARGS, "Evaluate a Starlark expression"},
    {"exec", (PyCFunction) Starlark_exec, METH_VARARGS, "Execute Starlark code, modifying the global state"},
    {NULL} /* Sentinel */
};

/* Starlark object type */
static PyTypeObject StarlarkType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "pystarlark._lib.starlark_go.Starlark",
    .tp_doc = "Starlark interpreter",
    .tp_basicsize = sizeof(StarlarkObject),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = (initproc) Starlark_new,
    .tp_dealloc = (destructor) Starlark_dealloc,
    .tp_methods = Starlark_methods
};

/* Module */
static PyModuleDef starlark_go = {
    PyModuleDef_HEAD_INIT,
    .m_name = "pystarlark._lib.starlark_go",
    .m_doc = "Interface to starlark-go",
    .m_size = -1,
};

/* Module initialization */
PyMODINIT_FUNC PyInit_starlark_go(void) {
    PyObject *m;
    if (PyType_Ready(&StarlarkType) < 0)
        return NULL;

    m = PyModule_Create(&starlark_go);
    if (m == NULL)
        return NULL;

    Py_INCREF(&StarlarkType);
    if (PyModule_AddObject(m, "Starlark", (PyObject *) &StarlarkType) < 0)
        goto dead;

    StarlarkError = PyErr_NewExceptionWithDoc(
        "pystarlark.StarlarkError",
        "Unspecified Starlark error",
        NULL,
        NULL
    );

    if ((!StarlarkError) || (PyModule_AddObject(m, "StarlarkError", StarlarkError) < 0))
        goto dead;

    ((PyTypeObject *)StarlarkError)->tp_str = StarlarkError_str;
    if (!add_getset((PyTypeObject *)StarlarkError, StarlarkError_getset))
        goto dead;

    SyntaxError = PyErr_NewExceptionWithDoc(
        "pystarlark.SyntaxError",
        "Starlark syntax error",
        StarlarkError,
        NULL
    );

    if ((!SyntaxError) || (PyModule_AddObject(m, "SyntaxError", SyntaxError) < 0))
        goto dead;

    ((PyTypeObject *)SyntaxError)->tp_str = StarlarkError_str;
    if (!add_getset((PyTypeObject *)SyntaxError, SyntaxError_getset))
        goto dead;

    EvalError = PyErr_NewExceptionWithDoc(
        "pystarlark.EvalError",
        "Starlark evaluation error",
        StarlarkError,
        NULL
    );

    if ((!EvalError) || (PyModule_AddObject(m, "EvalError", EvalError) < 0))
        goto dead;

    ((PyTypeObject *)EvalError)->tp_str = StarlarkError_str;
    if (!add_getset((PyTypeObject *)EvalError, EvalError_getset))
        goto dead;

    return m;

dead:
    Py_CLEAR(EvalError);
    Py_CLEAR(SyntaxError);
    Py_CLEAR(StarlarkError);
    Py_DECREF(&StarlarkType);
    Py_DECREF(m);

    return NULL;
}
