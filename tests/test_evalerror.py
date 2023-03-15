import pytest

from starlark_go import EvalError, Starlark

STARLARK_SRC = """
def wrong():
  return 1 + "2"
"""


def test_raises_evalerror():
    s = Starlark()

    with pytest.raises(EvalError):
        s.eval('1 + "2"')

    with pytest.raises(EvalError):
        s.exec('1 + "2"')


def test_eval_attrs():
    s = Starlark()
    raised = False

    s.eval(STARLARK_SRC, filename="fake.star")

    try:
        s.eval('1 + "2"')
    except EvalError as e:
        assert hasattr(e, "error")
        assert isinstance(e.error, str)
        assert hasattr(e, "error_type")
        assert isinstance(e.error_type, str)
        assert e.error_type == "*starlark.EvalError"
        assert hasattr(e, "filename")
        assert isinstance(e.filename, str)
        assert e.filename == "fake.star"
        assert hasattr(e, "line")
        assert isinstance(e.line, int)
        assert e.line == 1
        assert hasattr(e, "column")
        assert isinstance(e.column, int)
        assert e.column == 2
        assert hasattr(e, "function_name")
        assert isinstance(e.function_name, str)
        assert e.function_name == "wrong"
        assert hasattr(e, "backtrace")
        assert isinstance(e.backtrace, str)
        raised = True

    assert raised
