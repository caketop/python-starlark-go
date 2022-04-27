def test_import_starlarkerror():
    from starlark_go import StarlarkError

    assert issubclass(StarlarkError, BaseException)


def test_import_syntaxerror():
    from starlark_go import StarlarkError, SyntaxError

    assert issubclass(SyntaxError, StarlarkError)


def test_import_evalerror():
    from starlark_go import EvalError, StarlarkError

    assert issubclass(EvalError, StarlarkError)
