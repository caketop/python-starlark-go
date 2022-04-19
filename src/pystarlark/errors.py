from typing import Any, Optional, Tuple

__all__ = ["StarlarkError", "SyntaxError", "EvalError"]


class StarlarkError(Exception):
    def __init__(self, error: str, error_type: Optional[str] = None, *extra_args: Any):
        super().__init__(error, error_type, *extra_args)
        self.error = error
        self.error_type = self.__class__.__name__ if error_type is None else error_type

    def __str__(self) -> str:
        return self.error


class SyntaxError(StarlarkError):
    def __init__(
        self,
        error: str,
        error_type: str,
        msg: str,
        filename: str,
        line: int,
        column: int,
    ):
        super().__init__(error, error_type, msg, filename, line, column)
        self.msg = msg
        self.filename = filename
        self.line = line
        self.column = column


class EvalError(StarlarkError):
    def __init__(self, error: str, error_type: str, backtrace: str):
        super().__init__(error, error_type, backtrace)
        self.backtrace = backtrace


class ResolveErrorItem:
    def __init__(self, msg: str, line: int, column: int):
        self.msg = msg
        self.line = line
        self.column = column


class ResolveError(StarlarkError):
    def __init__(
        self, error: str, error_type: str, errors: Tuple[ResolveErrorItem, ...]
    ):
        super().__init__(error, error_type, errors)
        self.errors = list(errors)


class ConversionError(StarlarkError):
    pass
