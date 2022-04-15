#define PY_SSIZE_T_CLEAN
#include <Python.h>

/* This stuff is in the Go file */
unsigned long NewThread();
void DestroyThread(unsigned long threadId);
char *Eval(unsigned long threadId, char *stmt);
int ExecFile(unsigned long threadId, char *data);
void FreeCString(char *s);

/* Custom exceptions */
static PyObject *StarlarkError = NULL;
static PyObject *SyntaxError = NULL;
static PyObject *EvalError = NULL;

/* Helpers to raise custom exceptions from Go */
void Raise_StarlarkError(const char *error, const char *error_type) {
    PyGILState_STATE gilstate = PyGILState_Ensure();
    PyObject *exc_args = Py_BuildValue("ss", error, error_type);
	PyErr_SetObject(StarlarkError, exc_args);
    Py_DECREF(exc_args);
    PyGILState_Release(gilstate);
}

void Raise_SyntaxError(const char *error, const char *error_type, const char *msg, const char *filename, const long line, const long column) {
    PyGILState_STATE gilstate = PyGILState_Ensure();
    PyObject *exc_args = Py_BuildValue("ssssll", error, error_type, msg, filename, line, column);
	PyErr_SetObject(SyntaxError, exc_args);
    Py_DECREF(exc_args);
    PyGILState_Release(gilstate);
}

void Raise_EvalError(const char *error, const char *error_type, const char *backtrace) {
    PyGILState_STATE gilstate = PyGILState_Ensure();
    PyObject *exc_args = Py_BuildValue("sss", error, error_type, backtrace);
	PyErr_SetObject(EvalError, exc_args);
    Py_DECREF(exc_args);
    PyGILState_Release(gilstate);
}

/* Starlark object */
typedef struct {
    PyObject_HEAD
    unsigned long starlark_thread;
} StarlarkGo;

/* Starlark object methods */
static PyObject* StarlarkGo_new(PyTypeObject *type, PyObject *args, PyObject *kwds) {
    StarlarkGo *self;
    self = (StarlarkGo *) type->tp_alloc(type, 0);

    if (self != NULL)
        self->starlark_thread = NewThread();

    return (PyObject *) self;
}

static void StarlarkGo_dealloc(StarlarkGo *self) {
    DestroyThread(self->starlark_thread);
    Py_TYPE(self)->tp_free((PyObject *) self);
}

static PyObject* StarlarkGo_eval(StarlarkGo *self, PyObject *args) {
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

static PyObject* StarlarkGo_exec(StarlarkGo *self, PyObject *args) {
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

static PyMethodDef StarlarkGo_methods[] = {
    {"eval", (PyCFunction) StarlarkGo_eval, METH_VARARGS, "Evaluate a Starlark expression"},
    {"exec", (PyCFunction) StarlarkGo_exec, METH_VARARGS, "Execute Starlark code, modifying the global state"},
    {NULL} /* Sentinel */
};

/* Starlark object type */
static PyTypeObject StarlarkGoType = {
    PyVarObject_HEAD_INIT(NULL, 0)
    .tp_name = "pystarlark._lib.StarlarkGo",
    .tp_doc = "Starlark interpreter",
    .tp_basicsize = sizeof(StarlarkGo),
    .tp_itemsize = 0,
    .tp_flags = Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE,
    .tp_new = (initproc) StarlarkGo_new,
    .tp_dealloc = (destructor) StarlarkGo_dealloc,
    .tp_methods = StarlarkGo_methods
};

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
        PyErr_Format(PyExc_RuntimeError, "pystarlark.errors.%s is not defined", name);

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
    if (PyModule_AddObject(m, "StarlarkGo", (PyObject *) &StarlarkGoType) < 0) {
        Py_DECREF(&StarlarkGoType);
        Py_DECREF(m);

        return NULL;
    }

    return m;
}
