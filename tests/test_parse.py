from pystarlark import Starlark


def test_parse_int():
    s = Starlark()

    x = s.eval("7")
    assert isinstance(x, int)
    assert x == 7

    x = s.eval("7", parse=True)
    assert isinstance(x, int)
    assert x == 7

    x = s.eval("7", parse=False)
    assert isinstance(x, str)
    assert x == "7"


def test_parse_float():
    s = Starlark()

    x = s.eval("7.7")
    assert isinstance(x, float)
    assert x == 7.7

    x = s.eval("7.7", parse=True)
    assert isinstance(x, float)
    assert x == 7.7

    x = s.eval("7.7", parse=False)
    assert isinstance(x, str)
    assert x == "7.7"


def test_parse_bool():
    s = Starlark()

    x = s.eval("True")
    assert isinstance(x, bool)
    assert x is True

    x = s.eval("True", parse=True)
    assert isinstance(x, bool)
    assert x is True

    x = s.eval("True", parse=False)
    assert isinstance(x, str)
    assert x == "True"


def test_parse_none():
    s = Starlark()

    x = s.eval("None")
    assert x is None

    x = s.eval("None", parse=True)
    assert x is None

    x = s.eval("None", parse=False)
    assert isinstance(x, str)
    assert x == "None"


def test_parse_str():
    s = Starlark()

    x = s.eval('"True"')
    assert isinstance(x, str)
    assert x == "True"

    x = s.eval('"True"', parse=True)
    assert isinstance(x, str)
    assert x == "True"

    x = s.eval('"True"', parse=False)
    assert isinstance(x, str)
    assert x == '"True"'


def test_parse_list():
    s = Starlark()

    x = s.eval('[4, 2, 0, "go"]')
    assert isinstance(x, list)
    assert x == [4, 2, 0, "go"]

    x = s.eval('[4, 2, 0, "go"]', parse=True)
    assert isinstance(x, list)
    assert x == [4, 2, 0, "go"]

    x = s.eval('[4, 2, 0, "go"]', parse=False)
    assert isinstance(x, str)
    assert x == '[4, 2, 0, "go"]'


def test_parse_dict():
    s = Starlark()

    x = s.eval('{"lamb": "little", "pickles": 3}')
    assert isinstance(x, dict)
    assert x == {"lamb": "little", "pickles": 3}

    x = s.eval('{"lamb": "little", "pickles": 3}', parse=True)
    assert isinstance(x, dict)
    assert x == {"lamb": "little", "pickles": 3}

    x = s.eval('{"lamb": "little", "pickles": 3}', parse=False)
    assert isinstance(x, str)
    assert x.startswith("{")
    assert x.endswith("}")
