import setuptools

with open("README.md", "r") as fh:
    long_description = fh.read()

setuptools.setup(
    name="pystarlark",
    version="0.0.1",
    author="Kevin Chung",
    author_email="kchung@nyu.edu",
    description="Python bindings for Starlark in Go",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/ColdHeat/pystarlark",
    packages=setuptools.find_packages(),
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
    python_requires=">=3.6",
    # I'm not sure what this value is supposed to be
    build_golang={"root": "github.com/ColdHeat/pystarlark"},
    ext_modules=[setuptools.Extension("pystarlark/starlark", ["starlark.go"])],
    setup_requires=["setuptools-golang==2.3.0", "cffi==1.14.3"],
)
