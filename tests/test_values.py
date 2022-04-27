import pytest

from starlark_go import Starlark, configure_starlark
from starlark_go.errors import ResolveError

NESTED = [{"one": (1, 1, 1), "two": [2, {"two": 2222.22}]}, ("a", "b", "c")]
NESTED_STR = '[{"one": (1, 1, 1), "two": [2, {"two": 2222.22}]}, ("a", "b", "c")]'


def test_int():
    s = Starlark()

    x = s.eval("7")
    assert isinstance(x, int)
    assert x == 7

    x = s.eval("7", convert=True)
    assert isinstance(x, int)
    assert x == 7

    x = s.eval("7", convert=False)
    assert isinstance(x, str)
    assert x == "7"

    # too big to fit in 64 bits
    x = s.eval("10000000000000000000")
    assert isinstance(x, int)
    assert x == 10000000000000000000


def test_float():
    s = Starlark()

    x = s.eval("7.7")
    assert isinstance(x, float)
    assert x == 7.7

    x = s.eval("7.7", convert=True)
    assert isinstance(x, float)
    assert x == 7.7

    x = s.eval("7.7", convert=False)
    assert isinstance(x, str)
    assert x == "7.7"


def test_bool():
    s = Starlark()

    x = s.eval("True")
    assert isinstance(x, bool)
    assert x is True

    x = s.eval("True", convert=True)
    assert isinstance(x, bool)
    assert x is True

    x = s.eval("True", convert=False)
    assert isinstance(x, str)
    assert x == "True"


def test_none():
    s = Starlark()

    x = s.eval("None")
    assert x is None

    x = s.eval("None", convert=True)
    assert x is None

    x = s.eval("None", convert=False)
    assert isinstance(x, str)
    assert x == "None"


def test_str():
    s = Starlark()

    x = s.eval('"True"')
    assert isinstance(x, str)
    assert x == "True"

    x = s.eval('"True"', convert=True)
    assert isinstance(x, str)
    assert x == "True"

    x = s.eval('"True"', convert=False)
    assert isinstance(x, str)
    assert x == '"True"'


def test_list():
    s = Starlark()

    x = s.eval('[4, 2, 0, "go"]')
    assert isinstance(x, list)
    assert x == [4, 2, 0, "go"]

    x = s.eval('[4, 2, 0, "go"]', convert=True)
    assert isinstance(x, list)
    assert x == [4, 2, 0, "go"]

    x = s.eval('[4, 2, 0, "go"]', convert=False)
    assert isinstance(x, str)
    assert x == '[4, 2, 0, "go"]'


def test_dict():
    s = Starlark()

    x = s.eval('{"lamb": "little", "pickles": 3}')
    assert isinstance(x, dict)
    assert x == {"lamb": "little", "pickles": 3}

    x = s.eval('{"lamb": "little", "pickles": 3}', convert=True)
    assert isinstance(x, dict)
    assert x == {"lamb": "little", "pickles": 3}

    x = s.eval('{"lamb": "little", "pickles": 3}', convert=False)
    assert isinstance(x, str)
    assert x.startswith("{")
    assert x.endswith("}")


def test_set():
    s = Starlark()

    configure_starlark(allow_set=False)
    with pytest.raises(ResolveError):
        s.eval("set((1, 2, 3))")

    configure_starlark(allow_set=True)
    x = s.eval("set((1, 2, 3))")
    assert isinstance(x, set)
    assert x == set((1, 2, 3))

    x = s.eval("set((1, 2, 3))", convert=True)
    assert isinstance(x, set)
    assert x == set((1, 2, 3))

    x = s.eval("set((1, 2, 3))", convert=False)
    assert isinstance(x, str)
    assert x.startswith("set(")
    assert x.endswith(")")


def test_bytes():
    s = Starlark()

    x = s.eval("b'dead0000beef'")
    assert isinstance(x, bytes)
    assert x == b"dead0000beef"

    x = s.eval("b'dead0000beef'", convert=True)
    assert isinstance(x, bytes)
    assert x == b"dead0000beef"

    x = s.eval("b'dead0000beef'", convert=False)
    assert isinstance(x, str)
    assert x == 'b"dead0000beef"'


def test_tuple():
    s = Starlark()

    x = s.eval("(13, 37)")
    assert isinstance(x, tuple)
    assert x == (13, 37)

    x = s.eval("(13, 37)", convert=True)
    assert isinstance(x, tuple)
    assert x == (13, 37)

    x = s.eval("(13, 37)", convert=False)
    assert isinstance(x, str)
    assert x == "(13, 37)"


def test_nested():
    s = Starlark()

    x = s.eval(NESTED_STR)
    assert x == NESTED
