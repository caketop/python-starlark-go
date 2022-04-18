from ast import literal_eval
from typing import Any, Optional

from pystarlark._lib import StarlarkGo
from pystarlark.errors import EvalError, StarlarkError, SyntaxError

__all__ = ["Starlark", "StarlarkError", "EvalError", "SyntaxError"]
__version__ = "0.0.2"


class Starlark(StarlarkGo):
    def eval(
        self,
        expr: str,
        filename: Optional[str] = None,
        parse: Optional[bool] = None,
    ) -> Any:
        kwargs = {}
        should_parse = True

        if filename is not None:
            kwargs["filename"] = filename
        if parse is not None:
            kwargs["parse"] = parse
            should_parse = bool(parse)

        response = super().eval(expr, **kwargs)
        if should_parse:
            response = literal_eval(response)

        return response
