import pytest

from pystarlark import Starlark, configure_starlark
from pystarlark.errors import ResolveError

NESTED = [{"one": (1, 1, 1), "two": [2, {"two": 2222.22}]}, ("a", "b", "c")]
NESTED_STR = '[{"one": (1, 1, 1), "two": [2, {"two": 2222.22}]}, ("a", "b", "c")]'


def test_int():
    s = Starlark()

    s.exec("x = 7")
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), int)
    assert s.get("x") == 7

    # too big to fit in 64 bits
    s.exec("y = 10000000000000000000")
    assert len(s.globals()) == 2
    assert sorted(s.globals()) == ["x", "y"]
    assert isinstance(s.get("x"), int)
    assert isinstance(s.get("y"), int)
    assert s.get("x") == 7
    assert s.get("y") == 10000000000000000000


def test_float():
    s = Starlark()

    s.exec("x = 7.7")
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), float)
    assert s.get("x") == 7.7


def test_bool():
    s = Starlark()

    s.exec("x = True")
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), bool)
    assert s.get("x") is True


def test_none():
    s = Starlark()

    s.exec("x = None")
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert s.get("x") is None


def test_str():
    s = Starlark()

    s.exec('x = "True"')
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), str)
    assert s.get("x") == "True"


def test_list():
    s = Starlark()

    s.exec('x = [4, 2, 0, "go"]')
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), list)
    assert s.get("x") == [4, 2, 0, "go"]


def test_dict():
    s = Starlark()

    s.exec('x = {"lamb": "little", "pickles": 3}')
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), dict)
    assert s.get("x") == {"lamb": "little", "pickles": 3}


def test_set():
    s = Starlark()

    configure_starlark(allow_set=False)
    with pytest.raises(ResolveError):
        s.exec("x = set((1, 2, 3))")

    configure_starlark(allow_set=True)
    s.exec("x = set((1, 2, 3))")
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), set)
    assert s.get("x") == set((1, 2, 3))


def test_bytes():
    s = Starlark()

    s.exec("x = b'dead0000beef'")
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), bytes)
    assert s.get("x") == b"dead0000beef"


def test_tuple():
    s = Starlark()

    s.exec("x = (13, 37)")
    assert len(s.globals()) == 1
    assert s.globals() == ["x"]
    assert isinstance(s.get("x"), tuple)
    assert s.get("x") == (13, 37)


def test_nested():
    s = Starlark()

    s.exec(f"x = {NESTED_STR}")
    assert s.globals() == ["x"]
    assert len(s.globals()) == 1
    assert s.get("x") == NESTED
