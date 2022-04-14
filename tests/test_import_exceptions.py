def test_import_starlarkerror():
    from pystarlark import StarlarkError

    assert issubclass(StarlarkError, BaseException)


def test_import_syntaxerror():
    from pystarlark import StarlarkError, SyntaxError

    assert issubclass(SyntaxError, StarlarkError)


def test_import_evalerror():
    from pystarlark import EvalError, StarlarkError

    assert issubclass(EvalError, StarlarkError)


def test_import_unexpectederror():
    from pystarlark import StarlarkError, UnexpectedError

    assert issubclass(UnexpectedError, StarlarkError)
