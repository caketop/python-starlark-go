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
        assert hasattr(e, "error")
        assert isinstance(e.error, str)
        assert hasattr(e, "error_type")
        assert isinstance(e.error_type, str)
        # assert e.error_type == "syntax.Error"
        assert hasattr(e, "msg")
        assert isinstance(e.msg, str)
        assert hasattr(e, "filename")
        assert isinstance(e.filename, str)
        # assert e.filename == "<eval>"
        assert hasattr(e, "line")
        assert isinstance(e.line, int)
        # assert e.line == 1
        assert hasattr(e, "column")
        assert isinstance(e.column, int)
        # assert e.column == 2
        raised = True

    assert raised
