import pytest

from pystarlark import ResolveError, Starlark

ADD_ONE = """
def add_one(x):
    return x + 1
"""

ADD_TWO = """
def add_two(x):
    return add_one(add_one(x))
"""


def test_multi_exec():
    s = Starlark()

    s.exec(ADD_ONE)

    assert s.eval("add_one(1)") == 2

    with pytest.raises(ResolveError):
        s.eval("add_two(1)")

    s.exec(ADD_TWO)

    assert s.eval("add_two(1)") == 3
