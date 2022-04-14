#define PY_SSIZE_T_CLEAN
#include <Python.h>

unsigned long NewThread();
void DestroyThread(unsigned long threadId);
char *Eval(unsigned long threadId, char *stmt);
int ExecFile(unsigned long threadId, char *data);
void FreeCString(char *s);

static PyObject *StarlarkError;
static PyObject *EvalError;


typedef struct {
    PyObject_HEAD
    unsigned long starlark_thread;
} StarlarkObject;


void PyErr_Format_S(PyObject *exception, const char *message) {
    PyErr_Format(exception, message);
}

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
    if (PyModule_AddObject(m, "Starlark", (PyObject *) &StarlarkType) < 0) {
        Py_DECREF(&StarlarkType);
        Py_DECREF(m);
        return NULL;
    }

    StarlarkError = PyErr_NewExceptionWithDoc(
        "pystarlark.StarlarkError",
        "Base exception class for the pystarlark module",
        NULL,
        NULL
    );

    if ((!StarlarkError) || (PyModule_AddObject(m, "StarlarkError", StarlarkError) < 0)) {
        Py_CLEAR(StarlarkError);
        Py_DECREF(&StarlarkType);
        Py_DECREF(m);
        return NULL;
    }

    EvalError = PyErr_NewExceptionWithDoc(
        "pystarlark.EvalError",
        "Starlark evaluation error",
        StarlarkError,
        NULL
    );

    if ((!EvalError) || (PyModule_AddObject(m, "EvalError", EvalError) < 0)) {
        Py_CLEAR(EvalError);
        Py_CLEAR(StarlarkError);
        Py_DECREF(&StarlarkType);
        Py_DECREF(m);
        return NULL;
    }

    return m;
}
