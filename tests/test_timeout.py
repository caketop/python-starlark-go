import pytest

from starlark_go import EvalError, Starlark, configure_starlark

INFINITE_LOOP = """
def loop():
    while True:
        pass
loop()
"""


def test_exec_timeout():
    configure_starlark(allow_recursion=True)
    s = Starlark()
    with pytest.raises(EvalError, match="timed out"):
        s.exec(INFINITE_LOOP, timeout=0.5)


def test_eval_timeout():
    configure_starlark(allow_recursion=True)
    s = Starlark()
    s.exec("def loop():\n    while True:\n        pass")
    with pytest.raises(EvalError, match="timed out"):
        s.eval("loop()", timeout=0.5)


def test_exec_no_timeout_by_default():
    s = Starlark()
    s.exec("x = 1 + 2")
    assert s.get("x") == 3


def test_eval_no_timeout_by_default():
    s = Starlark()
    assert s.eval("1 + 2") == 3
