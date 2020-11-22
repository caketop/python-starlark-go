import json
from ast import literal_eval

from pystarlark.starlark import ffi, lib


class Starlark:
    def __init__(self):
        self._id = lib.NewThread()

    def read(self, code):
        if isinstance(code, str):
            code = code.encode()
        lib.ExecFile(self._id, code)

    def eval(self, statement, _raw=False):
        if isinstance(statement, str):
            statement = statement.encode()

        response = ffi.string(lib.Eval(self._id, statement))
        if _raw:
            return response
        value = json.loads(response)["value"]
        return literal_eval(value)

    def __del__(self):
        lib.DestroyThread(self._id)
