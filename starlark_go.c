#define PY_SSIZE_T_CLEAN
#include <Python.h>

unsigned long NewThread();
void DestroyThread(unsigned long threadId);
char *Eval(unsigned long threadId, char *stmt);
int ExecFile(unsigned long threadId, char *data);
void FreeCString(char *s);

static PyObject *StarlarkError = NULL;
static PyObject *SyntaxError = NULL;
static PyObject *EvalError = NULL;
static PyObject *UnexpectedError = NULL;


typedef struct {
    PyObject_HEAD
    unsigned long starlark_thread;
} StarlarkObject;


void Raise_SyntaxError(const char *error, const char *filename, const long line, const long column) {
    PyObject *exc_args = PyTuple_New(4);

	PyTuple_SetItem(exc_args, 0, PyUnicode_FromString(error));
	PyTuple_SetItem(exc_args, 1, PyUnicode_FromString(filename));
	PyTuple_SetItem(exc_args, 2, PyLong_FromLong(line));
	PyTuple_SetItem(exc_args, 3, PyLong_FromLong(column));
	PyErr_SetObject(SyntaxError, exc_args);
}

void Raise_EvalError(const char *error, const char *backtrace) {
    PyObject *exc_args = PyTuple_New(2);

	PyTuple_SetItem(exc_args, 0, PyUnicode_FromString(error));
	PyTuple_SetItem(exc_args, 1, PyUnicode_FromString(backtrace));
	PyErr_SetObject(EvalError, exc_args);
}

void Raise_UnexpectedError(const char *error, const char *err_type) {
    PyObject *exc_args = PyTuple_New(2);

	PyTuple_SetItem(exc_args, 0, PyUnicode_FromString(error));
	PyTuple_SetItem(exc_args, 1, PyUnicode_FromString(err_type));
	PyErr_SetObject(UnexpectedError, exc_args);
}

static PyObject *get_arg(PyObject *self, int index) {
    if (!PyObject_HasAttrString(self, "args"))
        return NULL;

	PyObject *args = PyObject_GetAttrString(self, "args");
    PyObject *arg = NULL;

    if (PyTuple_Size(args) > index)
        arg = PyTuple_GetItem(args, index);

    Py_XDECREF(args);
    return arg;
}

static PyObject *get_arg_str(PyObject *self, int index) {
    PyObject *arg = get_arg(self, index);
    if (arg != NULL)
        arg = PyObject_Str(arg);

    return arg;
}

static PyObject *get_arg_int(PyObject *self, int index) {
    PyObject *arg = get_arg(self, index);
    if ((arg != NULL) && (!PyLong_Check(arg)))
        return NULL;

    return arg;
}

int add_getset(PyTypeObject *type, PyGetSetDef *getsets) {
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

static PyObject *StarlarkError_message(PyObject *self) {
    PyObject *message = get_arg_str(self, 0);
    if (message == NULL)
        return PyUnicode_FromString("Unknown error");

    return message;
}

static PyObject *StarlarkError_second_string(PyObject *self) {
    PyObject *message = get_arg_str(self, 1);
    if (message == NULL)
        return PyUnicode_FromString("Unknown");

    return message;
}

static PyObject *StarlarkError_get_message(PyObject *self, void *closure) {
    return StarlarkError_message(self);
}

static PyObject *StarlarkError_get_second_string(PyObject *self, void *closure) {
    return StarlarkError_second_string(self);
}

static PyObject *SyntaxError_get_line(PyObject *self, void *closure) {
    PyObject *ret = get_arg_int(self, 2);
    if (ret == NULL)
        return PyLong_FromLong(-1);

    return ret;
}

static PyObject *SyntaxError_get_column(PyObject *self, void *closure) {
    PyObject *ret = get_arg_int(self, 3);
    if (ret == NULL)
        return PyLong_FromLong(-1);

    return ret;
}

static PyGetSetDef StarlarkError_getset[] = {
    {"message", StarlarkError_get_message, NULL, "A string containing an error message", NULL},
    {NULL}
};

static PyGetSetDef SyntaxError_getset[] = {
    {"filename", StarlarkError_get_second_string, NULL, "The name of the file in which the error occurred", NULL},
    {"line", SyntaxError_get_line, NULL, "The line on which the error occurred", NULL},
    {"column", SyntaxError_get_column, NULL, "The column on which the error occurred", NULL},
    {NULL}
};

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

    if (PyArg_ParseTuple(args, "U", &obj) == 0)
        return NULL;

    stmt = PyUnicode_AsUTF8String(obj);
    if (stmt == NULL)
        return NULL;

    cvalue = Eval(self->starlark_thread, PyBytes_AsString(stmt));

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

static PyModuleDef starlark_go = {
    PyModuleDef_HEAD_INIT,
    .m_name = "pystarlark._lib.starlark_go",
    .m_doc = "Interface to starlark-go",
    .m_size = -1,
};

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
        "Base exception class for the pystarlark module",
        NULL,
        NULL
    );

    if ((!StarlarkError) || (PyModule_AddObject(m, "StarlarkError", StarlarkError) < 0))
        goto dead;

    ((PyTypeObject *)StarlarkError)->tp_str = StarlarkError_message;
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

    ((PyTypeObject *)SyntaxError)->tp_str = StarlarkError_message;
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

    ((PyTypeObject *)EvalError)->tp_str = StarlarkError_message;

    UnexpectedError = PyErr_NewExceptionWithDoc(
        "pystarlark.UnexpectedError",
        "Unexpected error during Starlark evaluation",
        StarlarkError,
        NULL
    );

    if ((!UnexpectedError) || (PyModule_AddObject(m, "UnexpectedError", UnexpectedError) < 0))
        goto dead;

    ((PyTypeObject *)UnexpectedError)->tp_str = StarlarkError_message;

    return m;

dead:
    Py_CLEAR(UnexpectedError);
    Py_CLEAR(EvalError);
    Py_CLEAR(SyntaxError);
    Py_CLEAR(StarlarkError);
    Py_DECREF(&StarlarkType);
    Py_DECREF(m);

    return NULL;
}
