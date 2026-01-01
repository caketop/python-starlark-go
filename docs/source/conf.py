# Configuration file for the Sphinx documentation builder.
#
# This file only contains a selection of the most common options. For a full
# list see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Path setup --------------------------------------------------------------

# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use os.path.abspath to make it absolute, like shown here.
#
# import os
# import sys
# sys.path.insert(0, os.path.abspath('.'))

from importlib.metadata import version  # type: ignore

release = version("starlark_go")
version = ".".join(release.split(".")[:2])

# -- Project information -----------------------------------------------------

project = "python-starlark-go"
copyright = "2023, Jordan Webb"
author = "Jordan Webb"


# -- General configuration ---------------------------------------------------

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.
extensions = [
    "sphinx.ext.autodoc",
    "sphinx.ext.coverage",
    "myst_parser",
    "sphinx_rtd_theme",
    "sphinx.ext.intersphinx",
    "sphinx_search.extension",
]

# Add any paths that contain templates here, relative to this directory.
templates_path = ["_templates"]

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
# This pattern also affects html_static_path and html_extra_path.
exclude_patterns = ["_build", "Thumbs.db", ".DS_Store", "README.md"]

# Allow linking to Python docs
intersphinx_mapping = {"python": ("https://docs.python.org/3", None)}

# -- Options for HTML output -------------------------------------------------

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.
#
html_theme = "sphinx_rtd_theme"

html_theme_options = {
    "home_page_in_toc": True,
    "repository_url": "https://github.com/caketop/python-starlark-go",
    "repository_branch": "main",
    "use_repository_button": True,
    "use_issues_button": True,
}

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".
html_static_path = ["_static"]

manpages_url = "https://manpages.debian.org/{path}"

autodoc_default_options = {"member-order": "groupwise"}
autodoc_preserve_defaults = True
