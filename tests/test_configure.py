import pytest

from starlark_go import EvalError, Starlark, configure_starlark

RFIB = """
def rfib(n):
   if n <= 1:
       return n
   else:
       return(rfib(n-1) + rfib(n-2))

def fibonacci(n):
    r = []
    for i in range(n):
        r.append(rfib(i))
    return r
"""


def test_recursion():
    s = Starlark()
    s.exec(RFIB)

    configure_starlark(allow_recursion=False)
    with pytest.raises(EvalError):
        s.eval("fibonacci(5)")

    configure_starlark(allow_recursion=True)
    assert s.eval("fibonacci(5)") == [0, 1, 1, 2, 3]
    assert s.eval("fibonacci(10)") == [0, 1, 1, 2, 3, 5, 8, 13, 21, 34]


def test_retention():
    s = Starlark()

    configure_starlark(allow_set=True, allow_recursion=False)
    assert s.eval("set((1, 2, 3))") == set((1, 2, 3))

    # test that allow_set is untouched after setting a different value
    configure_starlark(allow_recursion=True)
    assert s.eval("set((1, 2, 3))") == set((1, 2, 3))
