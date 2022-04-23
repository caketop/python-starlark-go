from io import StringIO

import pytest

from pystarlark import Starlark


def test_print_eval(capsys: pytest.CaptureFixture[str]):
    s = Starlark()
    s.eval('print("hello")')

    captured = capsys.readouterr()
    assert captured.out == "hello\n"


def test_print_exec(capsys: pytest.CaptureFixture[str]):
    s = Starlark()
    s.exec('print("hello")')

    captured = capsys.readouterr()
    assert captured.out == "hello\n"


def test_more_eval():
    a = StringIO()
    b = StringIO()

    s = Starlark(print=lambda x: a.write(x + "\n"))
    s.eval('print("hello")')

    assert a.getvalue() == "hello\n"

    s.eval('print("hello")', print=lambda x: b.write(x + "\n"))

    assert a.getvalue() == "hello\n"
    assert b.getvalue() == "hello\n"

    s.print = lambda x: b.write(x + "\n")

    s.eval('print("goodbye")')

    assert a.getvalue() == "hello\n"
    assert b.getvalue() == "hello\ngoodbye\n"


def test_more_exec():
    a = StringIO()
    b = StringIO()

    s = Starlark(print=lambda x: a.write(x + "\n"))
    s.exec('print("hello")')

    assert a.getvalue() == "hello\n"

    s.exec('print("hello")', print=lambda x: b.write(x + "\n"))

    assert a.getvalue() == "hello\n"
    assert b.getvalue() == "hello\n"

    s.print = lambda x: b.write(x + "\n")

    s.exec('print("goodbye")')

    assert a.getvalue() == "hello\n"
    assert b.getvalue() == "hello\ngoodbye\n"
