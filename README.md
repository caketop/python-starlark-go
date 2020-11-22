# pystarlark

[![PyPI](https://img.shields.io/pypi/v/pystarlark)](https://pypi.org/project/pystarlark/)

Experimental Python bindings for [starlark-go](https://github.com/google/starlark-go)

## Installation

```
pip install pystarlark
```

## Examples

```python
from pystarlark import Starlark

s = Starlark()
fibonacci = """
def fibonacci(n=10):
   res = list(range(n))
   for i in res[2:]:
       res[i] = res[i-2] + res[i-1]
   return res
"""
s.read(fibonacci)
s.eval("fibonacci(5)")  # [0, 1, 1, 2, 3]
```

## How does this work?

pystarlark is a binding to [starlark-go](https://github.com/google/starlark-go) through a shared library built through cgo.

## What is Starlark?

Starlark is a Python-like language created by Google for its build system, Bazel. Starlark, while similar to Python, has some features that Python does not. Copied from the [main Starlark repo](https://github.com/bazelbuild/starlark#design-principles):

>  ## Design Principles
>
> *   **Deterministic evaluation**. Executing the same code twice will give the
>     same results.
> *   **Hermetic execution**. Execution cannot access the file system, network,
>     system clock. It is safe to execute untrusted code.
> *   **Parallel evaluation**. Modules can be loaded in parallel. To guarantee a
>     thread-safe execution, shared data becomes immutable.
> *   **Simplicity**. We try to limit the number of concepts needed to understand
>     the code. Users should be able to quickly read and write code, even if they
>     are not expert. The language should avoid pitfalls as much as possible.
> *   **Focus on tooling**. We recognize that the source code will be read,
>     analyzed, modified, by both humans and tools.
> *   **Python-like**. Python is a widely used language. Keeping the language
>     similar to Python can reduce the learning curve and make the semantics more
>     obvious to users.

In the words of the Starlark developers:

> Starlark is a dialect of Python. Like Python, it is a dynamically typed language with high-level data types, first-class functions with lexical scope, and garbage collection. Independent Starlark threads execute in parallel, so Starlark workloads scale well on parallel machines. Starlark is a small and simple language with a familiar and highly readable syntax. You can use it as an expressive notation for structured data, defining functions to eliminate repetition, or you can use it to add scripting capabilities to an existing application.

> A Starlark interpreter is typically embedded within a larger application, and the application may define additional domain-specific functions and data types beyond those provided by the core language.

## Why would I use this instead of just Python?

### Sandboxing

The primary reason this was written is for the "hermetic execution" feature of Starlark. Python is notoriously difficult to sandbox and there didn't appear to be any sandboxing solutions that could run within Python to run Python or Python-like code. While Starlark isn't exactly Python it is very very close to it. You can think of this as a secure way to run very simplistic Python functions. Note that this library itself doesn't really provide any security guarantees and your program may crash while using it (PRs welcome). Starlark itself is providing the security guarantees.

### Similar Work

[RestrictedPython](https://github.com/zopefoundation/RestrictedPython) looks pretty good and would probably work for most use cases including the one that pystarlark was written for. However, Python is notoriously difficult to sandbox and the developers of RestrictedPython even admit [that it causes headaches](https://docs.plone.org/develop/plone/security/sandboxing.html).

The [PyPy sandbox](https://doc.pypy.org/en/latest/sandbox.html) would probably work as a secure sandbox but it historically has been unmaintained and unsupported. While some significant work has recently gone into the sandbox, it primarily exists in a seperate branch in the PyPy repo. Also PyPy is a very heavy dependency to bring in if you're already running a Python interpreter.
