import pytest

from pystarlark import EvalError, Starlark


def test_raises_evalerror():
    s = Starlark()

    with pytest.raises(EvalError):
        s.eval('1 + "2"')

    with pytest.raises(EvalError):
        s.exec('1 + "2"')


def test_eval_attrs():
    s = Starlark()
    raised = False

    try:
        s.eval('1 + "2"')
    except EvalError as e:
        assert hasattr(e, "error")
        assert isinstance(e.error, str)
        assert hasattr(e, "error_type")
        assert isinstance(e.error_type, str)
        assert e.error_type == "*starlark.EvalError"
        assert hasattr(e, "backtrace")
        assert isinstance(e.backtrace, str)
        raised = True

    assert raised
