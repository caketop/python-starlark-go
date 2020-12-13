import json
from ast import literal_eval

from pystarlark.starlark import ffi, lib

__version__ = "0.0.2"


class Starlark:
    def __init__(self):
        self._id = lib.NewThread()

    def exec(self, code):
        if isinstance(code, str):
            code = code.encode()
        lib.ExecFile(self._id, code)

    def eval(self, statement, _raw=False):
        if isinstance(statement, str):
            statement = statement.encode()

        output = lib.Eval(self._id, statement)
        response = ffi.string(output)
        lib.FreeCString(output)
        if _raw:
            return response
        value = json.loads(response)["value"]
        return literal_eval(value)

    def __del__(self):
        lib.DestroyThread(self._id)
