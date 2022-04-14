import json

from ast import literal_eval
from typing import Any

from pystarlark._lib.starlark_go import EvalError
from pystarlark._lib.starlark_go import Starlark as LibStarlark
from pystarlark._lib.starlark_go import StarlarkError, SyntaxError, UnexpectedError

__all__ = ["Starlark", "StarlarkError", "EvalError", "SyntaxError", "UnexpectedError"]
__version__ = "0.0.2"


class Starlark(LibStarlark):
    def eval(self, statement: str, _raw: bool = False) -> Any:
        response = super().eval(statement)
        if _raw:
            return response
        value = json.loads(response)["value"]
        return literal_eval(value)
