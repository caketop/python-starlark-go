import pytest

from starlark_go import ConversionToPythonFailed, Starlark

STARLARK_SRC = """
foo = print
bar = [1, print, 2]
baz = {"c": print}
"""


DONT_KNOW = "Don't know how to convert Starlark *starlark.Builtin to Python"
LIST_INDEX_1 = (
    "While converting value <built-in function print> at index 1 in Starlark list: "
)
DICT_KEY_C = (
    'While converting value <built-in function print> of key "c" in Starlark dict: '
)


@pytest.fixture
def s() -> Starlark:
    starlark = Starlark()
    starlark.exec(STARLARK_SRC, filename="fake.star")
    return starlark


def test_ConversionToPythonFailed(s: Starlark):
    with pytest.raises(ConversionToPythonFailed) as e:
        s.eval("print")
    assert str(e.value) == DONT_KNOW

    s.exec("foo = print")

    with pytest.raises(ConversionToPythonFailed) as e:
        s.get("foo")
    assert str(e.value) == DONT_KNOW


def test_ConversionToPythonFailed_bar(s: Starlark):
    with pytest.raises(ConversionToPythonFailed) as e:
        s.eval("bar")
    assert str(e.value) == LIST_INDEX_1 + DONT_KNOW


def test_ConversionToPythonFailed_baz(s: Starlark):
    with pytest.raises(ConversionToPythonFailed) as e:
        s.eval("baz")
    assert str(e.value) == DICT_KEY_C + DONT_KNOW
