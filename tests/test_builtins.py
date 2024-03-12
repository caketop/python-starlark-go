import pytest

from starlark_go import Starlark, configure_starlark
from starlark_go.errors import ResolveError

NESTED = [{"one": (1, 1, 1), "two": [2, {"two": 2222.22}]}, ("a", "b", "c")]
NESTED_STR = '[{"one": (1, 1, 1), "two": [2, {"two": 2222.22}]}, ("a", "b", "c")]'


def test_builtin():
    def sum(a, *, b):
        return a + b
    
    s = Starlark(globals={
        "sum": sum
    })

    x = s.eval("sum(1, b=2)")
    assert isinstance(x, int)
    assert x == 3
