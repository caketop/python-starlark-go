#!/usr/bin/python
from cffi import FFI

ffibuilder = FFI()

ffibuilder.set_source(
    "pystarlark",
    """ //passed to the real C compiler
        #include "starlark.h"
    """,
    extra_objects=["starlark.so"],
)

ffibuilder.cdef(
    """
    extern void Hello();

    extern char* ExecCall(char* p0, char* p1);

    extern char* ExecCallEval(char* p0, char* p1);

    extern char* ExecEval(char* p0);
    """
)

if __name__ == "__main__":
    ffibuilder.compile(verbose=True)
