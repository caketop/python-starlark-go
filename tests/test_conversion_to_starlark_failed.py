from typing import Sequence

import pytest

from starlark_go import ConversionToStarlarkFailed, Starlark

foo = print
bar = [1, print, 2]
baz = {"c": print}


DONT_KNOW = "Don't know how to convert Python builtin_function_or_method to Starlark"
LIST_INDEX_1 = "While converting value at index 1 in Python list: "
DICT_KEY_C = 'While converting value of key "c" in Python dict: '


class ExplodingException(Exception):
    pass


class ExplodingSequence(Sequence[str]):
    def __len__(self):
        return 3

    def __getitem__(self, key: int):
        if key == 1:
            raise ExplodingException("Surprise!")
        return key * 100


@pytest.fixture
def s() -> Starlark:
    starlark = Starlark()
    return starlark


def test_ConversionToStarlarkFailed(s: Starlark):
    with pytest.raises(ConversionToStarlarkFailed) as e:
        s.set(foo=foo)
    assert str(e.value) == DONT_KNOW
    assert getattr(e.value, "__cause__", None) is None


def test_ConversionToStarlarkFailed_bar(s: Starlark):
    with pytest.raises(ConversionToStarlarkFailed) as e:
        s.set(bar=bar)
    assert str(e.value) == LIST_INDEX_1 + DONT_KNOW
    assert getattr(e.value, "__cause__", None) is None


def test_ConversionToStarlarkFailed_baz(s: Starlark):
    with pytest.raises(ConversionToStarlarkFailed) as e:
        s.set(baz=baz)
    assert str(e.value) == DICT_KEY_C + DONT_KNOW
    assert getattr(e.value, "__cause__", None) is None


def test_exploding_sequence(s: Starlark):
    with pytest.raises(ConversionToStarlarkFailed) as e:
        s.set(surprise=ExplodingSequence())
    print(str(e.value))
    assert isinstance(e.value, ConversionToStarlarkFailed)
    assert hasattr(e.value, "__cause__")
    assert isinstance(getattr(e.value, "__cause__"), ExplodingException)
