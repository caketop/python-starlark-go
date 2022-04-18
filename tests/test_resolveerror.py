from pystarlark import ResolveError, Starlark


def test_eval_resolveerror():
    s = Starlark()
    raised = False

    try:
        s.eval("add_one(1)")
    except ResolveError as e:
        assert isinstance(e.errors, list)
        assert len(e.errors) == 1
        assert e.errors[0].line == 1
        assert e.errors[0].column == 1
        assert e.errors[0].msg == "undefined: add_one"
        raised = True

    try:
        s.eval("from_bad(True) + to_worse(True)")
    except ResolveError as e:
        assert isinstance(e.errors, list)
        assert len(e.errors) == 2
        assert e.errors[0].line == 1
        assert e.errors[0].column == 1
        assert e.errors[0].msg == "undefined: from_bad"
        assert e.errors[1].line == 1
        assert e.errors[1].column == 18
        assert e.errors[1].msg == "undefined: to_worse"
        raised = True

    assert raised


def test_exec_resolveerror():
    s = Starlark()
    raised = False

    try:
        s.exec("add_one(1)")
    except ResolveError as e:
        assert isinstance(e.errors, list)
        assert len(e.errors) == 1
        assert e.errors[0].line == 1
        assert e.errors[0].column == 1
        assert e.errors[0].msg == "undefined: add_one"
        raised = True

    try:
        s.exec("from_bad(True) + to_worse(True)")
    except ResolveError as e:
        assert isinstance(e.errors, list)
        assert len(e.errors) == 2
        assert e.errors[0].line == 1
        assert e.errors[0].column == 1
        assert e.errors[0].msg == "undefined: from_bad"
        assert e.errors[1].line == 1
        assert e.errors[1].column == 18
        assert e.errors[1].msg == "undefined: to_worse"
        raised = True

    assert raised
