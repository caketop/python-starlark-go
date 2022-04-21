from setuptools import Extension, setup

# I would only use setup.cfg but it can't compile extensions, so here we are.

setup(
    build_golang={"root": "github.com/caketop/pystarlark"},
    ext_modules=[Extension("pystarlark/starlark_go", ["python_object.go"])],
)
