import json
from ast import literal_eval
from typing import Any

from pystarlark._lib.starlark_go import Starlark as BaseStarlark

__version__ = "0.0.2"


class Starlark(BaseStarlark):
    def eval(self, statement: str, _raw: bool = False) -> Any:
        response = super().eval(statement)
        if _raw:
            return response
        value = json.loads(response)["value"]
        return literal_eval(value)
