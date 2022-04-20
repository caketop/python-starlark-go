import pytest

from pystarlark import Starlark, configure_starlark
from pystarlark.errors import ResolveError

NESTED = [{"one": (1, 1, 1), "two": [2, {"two": 2222.22}]}, ("a", "b", "c")]
NESTED_STR = '[{"one": (1, 1, 1), "two": [2, {"two": 2222.22}]}, ("a", "b", "c")]'


def test_int():
    s = Starlark()

    s.exec("x = 7")
    assert len(s) == 1
    assert s.keys() == ["x"]
    assert "x" in s
    assert isinstance(s["x"], int)
    assert s["x"] == 7

    # too big to fit in 64 bits
    s.exec("y = 10000000000000000000")
    assert len(s) == 2
    assert sorted(s.keys()) == ["x", "y"]
    assert "x" in s
    assert "y" in s
    assert isinstance(s["x"], int)
    assert isinstance(s["y"], int)
    assert s["x"] == 7
    assert s["y"] == 10000000000000000000


def test_float():
    s = Starlark()

    s.exec("x = 7.7")
    assert len(s) == 1
    assert s.keys() == ["x"]
    assert "x" in s
    assert isinstance(s["x"], float)
    assert s["x"] == 7.7


def test_bool():
    s = Starlark()

    s.exec("x = True")
    assert len(s) == 1
    assert s.keys() == ["x"]
    assert "x" in s
    assert isinstance(s["x"], bool)
    assert s["x"] is True


def test_none():
    s = Starlark()

    s.exec("x = None")
    assert len(s) == 1
    assert s.keys() == ["x"]
    assert "x" in s
    assert s["x"] is None


def test_str():
    s = Starlark()

    s.exec('x = "True"')
    assert len(s) == 1
    assert s.keys() == ["x"]
    assert "x" in s
    assert isinstance(s["x"], str)
    assert s["x"] == "True"


def test_list():
    s = Starlark()

    s.exec('x = [4, 2, 0, "go"]')
    assert len(s) == 1
    assert s.keys() == ["x"]
    assert "x" in s
    assert isinstance(s["x"], list)
    assert s["x"] == [4, 2, 0, "go"]


def test_dict():
    s = Starlark()

    s.exec('x = {"lamb": "little", "pickles": 3}')
    assert len(s) == 1
    assert s.keys() == ["x"]
    assert "x" in s
    assert isinstance(s["x"], dict)
    assert s["x"] == {"lamb": "little", "pickles": 3}


def test_set():
    s = Starlark()

    configure_starlark(allow_set=False)
    with pytest.raises(ResolveError):
        s.exec("x = set((1, 2, 3))")

    configure_starlark(allow_set=True)
    s.exec("x = set((1, 2, 3))")
    assert len(s) == 1
    assert s.keys() == ["x"]
    assert "x" in s
    assert isinstance(s["x"], set)
    assert s["x"] == set((1, 2, 3))


def test_bytes():
    s = Starlark()

    s.exec("x = b'dead0000beef'")
    assert len(s) == 1
    assert s.keys() == ["x"]
    assert "x" in s
    assert isinstance(s["x"], bytes)
    assert s["x"] == b"dead0000beef"


def test_tuple():
    s = Starlark()

    s.exec("x = (13, 37)")
    assert len(s) == 1
    assert s.keys() == ["x"]
    assert "x" in s
    assert isinstance(s["x"], tuple)
    assert s["x"] == (13, 37)


def test_nested():
    s = Starlark()

    s.exec(f"x = {NESTED_STR}")
    assert s.keys() == ["x"]
    assert len(s) == 1
    assert "x" in s
    assert s["x"] == NESTED
