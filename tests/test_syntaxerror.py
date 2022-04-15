import pytest

from pystarlark import Starlark, SyntaxError


def test_raises_syntaxerror():
    s = Starlark()

    assert s.eval("7") == 7

    with pytest.raises(SyntaxError):
        s.eval(" 7 ")

    with pytest.raises(SyntaxError):
        s.exec(" 7 ")


def test_syntaxerror_attrs():
    s = Starlark()
    raised = False

    try:
        s.eval(" 7 ")
    except SyntaxError as e:
        assert hasattr(e, "message")
        assert isinstance(e.message, str)
        assert hasattr(e, "filename")
        assert isinstance(e.filename, str)
        assert hasattr(e, "line")
        assert isinstance(e.line, int)
        assert hasattr(e, "column")
        assert isinstance(e.column, int)
        raised = True

    assert raised
