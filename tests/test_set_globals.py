import sys

import pytest

from starlark_go import EvalError, Starlark, StarlarkError

NESTED = [{"one": (1, 1, 1), "two": [2, {"two": 2222.22}]}, ("a", "b", "c")]


def test_set_globals():
    s = Starlark()

    s.set()
    assert len(s.globals()) == 0

    s.set(x=1)
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]

    s.set(x=1, y=[2], z={3: 3})
    assert len(s.globals()) == 3
    assert sorted(s.globals()) == ["x", "y", "z"]

    s2 = Starlark(globals={"x": 1, "y": 2, "z": 3})
    assert len(s2.globals()) == 3
    assert sorted(s2.globals()) == ["x", "y", "z"]

    s3 = Starlark(globals={})
    assert len(s3.globals()) == 0

    with pytest.raises(TypeError):
        Starlark(globals=True)  # type: ignore

    with pytest.raises(TypeError):
        Starlark(globals=[1, 2, 3])  # type: ignore

    with pytest.raises(TypeError):
        Starlark(globals="nope")  # type: ignore

    with pytest.raises(TypeError):
        Starlark(globals=b"dead")  # type: ignore

    with pytest.raises(TypeError):
        Starlark(globals=set((1, 2, 3)))  # type: ignore


def test_int():
    s = Starlark()

    s.set(x=7)
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), int)
    assert s.get("x") == 7
    assert s.eval("x + 1") == 8

    # too big to fit in 64 bits
    s.set(y=10000000000000000000)
    assert len(s.globals()) == 2
    assert sorted(s.globals()) == ["x", "y"]
    assert isinstance(s.get("x"), int)
    assert isinstance(s.get("y"), int)
    assert s.get("x") == 7
    assert s.get("y") == 10000000000000000000
    assert s.eval("y + 1") == 10000000000000000001


def test_float():
    s = Starlark()

    s.set(x=7.7)
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), float)
    assert s.get("x") == 7.7
    assert s.eval("int(x)") == 7
    assert s.eval("int(x + 1)") == 8


def test_bool():
    s = Starlark()

    s.set(x=True)
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), bool)
    assert s.get("x") is True
    assert s.eval("not x") is False


def test_none():
    s = Starlark()

    s.set(x=None)
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert s.get("x") is None


def test_str():
    s = Starlark()

    s.set(x="True")
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), str)
    assert s.get("x") == "True"
    assert s.eval("x + 'True'") == "TrueTrue"


def test_list():
    s = Starlark()

    s.set(x=[4, 2, 0, "go"])
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), list)
    assert s.get("x") == [4, 2, 0, "go"]


def test_dict():
    s = Starlark()

    s.set(x={"lamb": "little", "pickles": 3})
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), dict)
    assert s.get("x") == {"lamb": "little", "pickles": 3}


def test_set():
    s = Starlark()

    s.set(x=set((1, 2, 3)))
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), set)
    assert s.get("x") == set((1, 2, 3))


def test_bytes():
    s = Starlark()

    s.set(x=b"dead0000beef")
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), bytes)
    assert s.get("x") == b"dead0000beef"


def test_tuple():
    s = Starlark()

    s.set(x=(13, 37))
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), tuple)
    assert s.get("x") == (13, 37)


def test_nested():
    s = Starlark()

    s.set(x=NESTED)
    assert s.globals() == ["x"]
    assert len(s.globals()) == 1
    assert s.get("x") == NESTED


def test_func():
    s = Starlark()

    def func_impl(x):
        if x == 0:
            raise ValueError("got zero")
        return x * 2

    s.set(func=func_impl)
    assert s.globals() == ["func"]

    with pytest.raises(
        StarlarkError,
        match=r"Don't know how to convert Starlark \*starlark.Builtin to Python",
    ):
        s.get("func")

    assert s.eval("func(10)") == 20
    assert s.eval("func(x = 10)") == 20

    with pytest.raises(EvalError, match=r"<builtin> in func_impl:0:0: got zero"):
        s.eval("func(0)")

    name = "func_impl" if sys.version_info < (3, 10) else "test_func.<locals>.func_impl"

    with pytest.raises(EvalError, match=r"<builtin> in func_impl:0:0: " + name + r"\(\) missing 1 required positional argument: 'x'"):
        s.eval("func()")

    with pytest.raises(EvalError, match=r"<builtin> in func_impl:0:0: " + name + r"\(\) got an unexpected keyword argument 'unknown'"):
        s.eval("func(unknown=0)")


def test_func_references():
    s = Starlark()

    def func_new_ref():
        return {"a": 1}

    def func_arg_ref(x):
        return {"a": 1, "b": x}

    s.set(
        func_new_ref=func_new_ref,
        func_arg_ref=func_arg_ref,
    )

    assert s.eval("func_new_ref()") == {"a": 1}
    assert s.eval("func_arg_ref([1, 2, 3])") == {"a": 1, "b": [1, 2, 3]}

def test_method():
    s = Starlark()

    class Test:
        def __init__(self):
            self.result = None

        def func_impl(self, x):
            if x == 0:
                raise ValueError("got zero")
            self.result = x * 2

    test = Test()
    s.set(func=test.func_impl)
    assert s.globals() == ["func"]

    with pytest.raises(
        StarlarkError,
        match=r"Don't know how to convert Starlark \*starlark.Builtin to Python",
    ):
        s.get("func")

    s.exec("func(10)")
    assert test.result == 20

    with pytest.raises(EvalError, match=r"<builtin> in func_impl:0:0: got zero"):
        s.exec("func(0)")

    name = "func_impl" if sys.version_info < (3, 10) else "test_method.<locals>.Test.func_impl"

    with pytest.raises(EvalError, match=r"<builtin> in func_impl:0:0: " + name + r"\(\) missing 1 required positional argument: 'x'"):
        s.eval("func()")

    with pytest.raises(EvalError, match=r"<builtin> in func_impl:0:0: " + name + r"\(\) got an unexpected keyword argument 'unknown'"):
        s.eval("func(unknown=0)")


def test_method_references():
    s = Starlark()

    class Test:
        def __init__(self):
            self.result = None

        def func_new_ref(self):
            self.result = {"a": 1}

        def func_arg_ref(self, x):
            self.result = {"a": 1, "b": x}

    test = Test()
    s.set(
        func_new_ref=test.func_new_ref,
        func_arg_ref=test.func_arg_ref,
    )

    s.exec("func_new_ref()")
    assert test.result == {"a": 1}
    s.exec("func_arg_ref([1, 2, 3])")
    assert test.result == {"a": 1, "b": [1, 2, 3]}
