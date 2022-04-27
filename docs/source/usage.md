# Usage

## Create a context

A {py:obj}`starlark_go.Starlark` object is needed to provide a context (mainly, a set of global variables) for executing Starlark code. Creating an empty one is simple:

```python
from starlark_go import Starlark

s = Starlark()
```

The `globals` keyword argument to the A {py:obj}`starlark_go.Starlark` constructor can be used to pass a dictionary containing some initial global variables:

```python
from starlark_go import Starlark

s = Starlark(globals={"a": 1, "b": 3})
```

## Evaluating code

{py:meth}`starlark_go.Starlark.eval` can be used to evaluate a Starlark expression:

```python
from starlark_go import Starlark

s = Starlark()

s.eval("2 + 2") # 4
```

Starlark syntax is more-or-less identical to Python. Expressions can reference variables, just like you might in Python:

```python
from starlark_go import Starlark

s = Starlark(globals={"a": 1, "b": 3})

s.eval("a + b") # 4
```

## Defining variables and functions

{py:meth}`starlark_go.Starlark.eval` is only for evaluating expressions; if you want to define things in Starlark, you'll need to use {py:meth}`starlark_go.Starlark.exec`.

```python
from starlark_go import Starlark

s = Starlark()

s.exec("a = 1")
s.exec("b = 3")
s.eval("a + b") # 4
```

There is no distinction between variables set by `globals` versus variables set by `exec`; it is simply another way to set a variable.

{py:meth}`starlark_go.Starlark.exec` can also be used to define functions in Starlark. Remember, Starlark's syntax is more-or-less identical to Python:

```python
from starlark_go import Starlark

s = Starlark()

s.exec("""
def add_one(x):
  return x + 1
""")

s.eval("add_one(3)") # 4
```

## Defining variables from Python

{py:meth}`starlark_go.Starlark.set` can be used to define one or more Starlark global variables:

```python
from starlark_go import Starlark

s = Starlark()

s.set(a=1, b=3)

s.eval("a + b") # 4
```

There is no distinction between variables set by `set` versus other variables; it is simply another way to set a variable.

## Retrieving variables

{py:meth}`starlark_go.Starlark.get` can be used to retrieve a Starlark global variable:

```python
from starlark_go import Starlark

s = Starlark(globals={"b": 3, "c": True})

s.exec("a = 1")
s.set(d=[1, 2, 3])

s.get("a") # 1
s.get("b") # 3
s.get("c") # True
s.get("d") # [1, 2, 3]
```

A default value can be provided to {py:meth}`starlark_go.Starlark.get`; if one is not provided and the variable you are attempting to retrieve does not exist, a KeyError will be raised:

```python
from starlark_go import Starlark

s = Starlark()

s.get("e") # !!! raises KeyError !!!
s.get("e", 72) # 72
```

## Removing variables

{py:meth}`starlark_go.Starlark.pop` functions identically to {py:meth}`starlark_go.Starlark.get`, except that it removes the variable before returning its value:

```python
from starlark_go import Starlark

s = Starlark(globals={"a": 1, "b": 2})

s.eval("a + b") # 4
s.pop("a") # 1
s.eval("a + b") # !!! raises ResolveError !!!
```

## Overriding the `print()` function

By default, Starlark's `print()` function is routed to Python's built-in {py:func}`python:print`, but you can provide a different function to override it.

This can be done when you create the context:

```python
import logging

from starlark_go import Starlark

s = Starlark(print=logging.warning)
```

...or after it is created:

```python
import logging

from starlark_go import Starlark

s = Starlark()
s.print = logging.warning
```

...or for individual calls to {py:meth}`starlark_go.Starlark.eval` and {py:meth}`starlark_go.Starlark.exec`:


```python
import logging

from starlark_go import Starlark

s = Starlark()
s.exec('print("hello!")', print=logging.warning)
```
