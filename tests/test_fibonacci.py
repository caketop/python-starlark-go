from pystarlark import Starlark


def test_fibonacci():
    s = Starlark()
    fibonacci = """
def fibonacci(n=10):
    res = list(range(n))
    for i in res[2:]:
        res[i] = res[i-2] + res[i-1]
    return res
"""
    s.exec(fibonacci)
    assert s.eval("fibonacci(5)") == [0, 1, 1, 2, 3]
    assert s.eval("fibonacci(10)") == [0, 1, 1, 2, 3, 5, 8, 13, 21, 34]
