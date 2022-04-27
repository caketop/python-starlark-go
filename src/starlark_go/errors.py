from typing import Any, Optional, Tuple

__all__ = ["StarlarkError", "SyntaxError", "EvalError"]


class StarlarkError(Exception):
    """
    Base class for Starlark errors.

    All Starlark-specific errors that are raised by :py:class:`starlark_go.Starlark`
    are derived from this class.
    """

    def __init__(self, error: str, error_type: Optional[str] = None, *extra_args: Any):
        super().__init__(error, error_type, *extra_args)
        self.error = error
        """
        A description of the error.

        :type: str
        """
        self.error_type = self.__class__.__name__ if error_type is None else error_type
        """
        The name of the Go type of the error.

        :type: typing.Optional[str]
        """

    def __str__(self) -> str:
        return self.error


class SyntaxError(StarlarkError):
    """
    A Starlark syntax error.

    This exception is raised when syntatically invalid code is passed to
    :py:meth:`starlark_go.Starlark.eval` or :py:meth:`starlark_go.Starlark.exec`.
    """

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
        """
        A description of the syntax error

        :type: str
        """
        self.filename = filename
        """
        The name of the file that the error occurred in (taken from the ``filename``
        parameter to :py:meth:`starlark_go.Starlark.eval`
        or :py:meth:`starlark_go.Starlark.exec`.)

        :type: str
        """
        self.line = line
        """
        The line number that the error occurred on (1-based)

        :type: int
        """
        self.column = column
        """
        The column that the error occurred on (1-based)

        :type: int
        """


class EvalError(StarlarkError):
    """
    A Starlark evaluation error.

    This exception is raised when otherwise valid code attempts an illegal operation,
    such as adding a string to an integer.
    """

    def __init__(self, error: str, error_type: str, backtrace: str):
        super().__init__(error, error_type, backtrace)
        self.backtrace = backtrace
        """
        A backtrace through Starlark's stack leading up to the error.

        :type: str
        """


class ResolveErrorItem:
    """
    A location associated with a :py:class:`ResolveError`.
    """

    def __init__(self, msg: str, line: int, column: int):
        self.msg = msg
        """
        A description of the problem at the location.

        :type: str
        """
        self.line = line
        """
        The line where the problem occurred (1-based)

        :type: int
        """
        self.column = column
        """
        The column where the problem occurred (1-based)

        :type: int
        """


class ResolveError(StarlarkError):
    """
    A Starlark resolution error.

    This exception is raised by
    :py:meth:`starlark_go.Starlark.eval` or :py:meth:`starlark_go.Starlark.exec`
    when an undefined name is referenced.
    """

    def __init__(
        self, error: str, error_type: str, errors: Tuple[ResolveErrorItem, ...]
    ):
        super().__init__(error, error_type, errors)
        self.errors = list(errors)
        """
        A list of locations where resolution errors occurred. A ResolveError may
        contain one or more locations.

        :type: typing.List[ResolveErrorItem]
        """


class ConversionError(StarlarkError):
    """
    A Starlark conversion error.

    This exception is raied by :py:meth:`starlark_go.Starlark.eval`,
    :py:meth:`starlark_go.Starlark.get`, and :py:meth:`starlark_go.Starlark.set`
    when a Starlark value can not be converted to a Python value, or when a
    Python value can not be converted to a Starlark value.
    """
