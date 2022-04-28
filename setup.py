from setuptools import Extension, setup

# I would only use setup.cfg but it can't compile extensions, so here we are.

setup(
    build_golang={"root": "github.com/caketop/python-starlark-go", "strip": False},
    ext_modules=[Extension("starlark_go/starlark_go", ["python_object.go"])],
)
