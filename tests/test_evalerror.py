import pytest
from pystarlark import EvalError, Starlark


def test_evalerror():
    s = Starlark()

    assert s.eval("7") == 7

    with pytest.raises(EvalError):
        s.eval(" 7 ")

    with pytest.raises(EvalError):
        s.exec(" 7 ")
