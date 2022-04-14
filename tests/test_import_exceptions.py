def test_import_starlarkerror():
    from pystarlark import StarlarkError

    assert issubclass(StarlarkError, BaseException)


def test_import_evalerror():
    from pystarlark import EvalError, StarlarkError

    assert issubclass(EvalError, StarlarkError)
