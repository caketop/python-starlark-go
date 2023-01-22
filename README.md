# python-starlark-go

[![PyPI](https://img.shields.io/pypi/v/starlark-go)](https://pypi.org/project/starlark-go/)
![Tests](https://github.com/caketop/python-starlark-go/actions/workflows/test.yml/badge.svg)
[![Documentation Status](https://readthedocs.org/projects/python-starlark-go/badge/?version=latest)](https://python-starlark-go.readthedocs.io/en/latest/?badge=latest)
[![action-bumpr supported](https://img.shields.io/badge/bumpr-supported-ff69b4?logo=github&link=https://github.com/haya14busa/action-bumpr)](https://github.com/haya14busa/action-bumpr)

## Introduction

This module provides Python bindings for the Starlark programming language.

Starlark is a dialect of Python designed for hermetic execution and deterministic evaluation. That means you can run Starlark code you don't trust without worrying about it being able access any data you did not explicitly supply to it, and that you can count on the same code to always produce the same value when used with the same input data. Starlark was orginally created for the [Bazel build system](https://bazel.build/). There are now several implementations of Starlark; this module is built on [starlark-go](https://github.com/google/starlark-go).

This module was originally forked from Kevin Chung's [pystarlark](https://github.com/ColdHeat/pystarlark).

The version of starlark-go that is currently embedded in this module is [v0.0.0-20220328144851-d1966c6b9fcd](https://pkg.go.dev/go.starlark.net@v0.0.0-20220328144851-d1966c6b9fcd).

## Features

- Evalutes Starlark code from Python
- Translates Starlark exceptions to Python exceptions
- Converts Starlark values to Python values
- Converts Python values to Starlark values
- No runtime dependencies - cgo is used to bundle [starlark-go](https://github.com/google/starlark-go) into a Python extension

## Installation

```
pip install starlark-go
```

## Usage

```python
from starlark_go import Starlark

s = Starlark()
fibonacci = """
def fibonacci(n=10):
   res = list(range(n))
   for i in res[2:]:
       res[i] = res[i-2] + res[i-1]
   return res
"""
s.exec(fibonacci)
s.eval("fibonacci(5)")  # [0, 1, 1, 2, 3]

s.set(x=5)
s.eval("x") # 5
s.eval("fibonacci(x)")  # [0, 1, 1, 2, 3]
```
