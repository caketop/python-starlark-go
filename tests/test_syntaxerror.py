import pytest

from pystarlark import Starlark, SyntaxError


def test_syntaxerror():
    s = Starlark()

    assert s.eval("7") == 7

    with pytest.raises(SyntaxError):
        s.eval(" 7 ")

    with pytest.raises(SyntaxError):
        s.exec(" 7 ")
