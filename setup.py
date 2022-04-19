import re

import setuptools

with open("src/pystarlark/__init__.py", "r", encoding="utf8") as f:
    version = re.search(r'__version__ = "(.*?)"', f.read()).group(1)

with open("README.md", "r") as fh:
    long_description = fh.read()

setuptools.setup(
    name="pystarlark",
    version=version,
    author="Kevin Chung",
    author_email="kchung@nyu.edu",
    description="Python bindings for Starlark in Go",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/ColdHeat/pystarlark",
    package_dir={"": "src"},
    packages=setuptools.find_packages("src"),
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    python_requires=">=3.7",
    # I'm not sure what this value is supposed to be
    build_golang={"root": "github.com/caketop/pystarlark"},
    package_data={"pystarlark": ["*.pyi", "py.typed"]},
    ext_modules=[setuptools.Extension("pystarlark/starlark_go", ["starlark.go"])],
    setup_requires=["setuptools-golang==2.7.0"],
)
