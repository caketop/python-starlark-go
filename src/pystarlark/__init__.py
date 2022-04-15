import json

from ast import literal_eval
from typing import Any

from pystarlark._lib import StarlarkGo
from pystarlark.errors import EvalError, StarlarkError, SyntaxError

__all__ = ["Starlark", "StarlarkError", "EvalError", "SyntaxError"]
__version__ = "0.0.2"


class Starlark(StarlarkGo):
    def eval(self, statement: str, _raw: bool = False) -> Any:
        response = super().eval(statement)
        if _raw:
            return response
        value = json.loads(response)["value"]
        return literal_eval(value)
