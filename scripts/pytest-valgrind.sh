#!/bin/sh

set -e

VALGRIND_ROOT=/tmp/valgrind-python
VALGRIND_PYTHON="$VALGRIND_ROOT/bin/python3"
VALGRIND_LOG="$VALGRIND_ROOT/valgrind.log"

set -x

# Set up a venv
if [ ! -d "$VALGRIND_ROOT" ]; then
  python3.10 -m venv "$VALGRIND_ROOT"
  "$VALGRIND_PYTHON" -m pip install pytest pytest-valgrind
fi

# Install the module
"$VALGRIND_PYTHON" -m pip install .

# Nasty hack - rebuild with all the debugging symbols
PYSTARLARK_SO="$VALGRIND_ROOT/lib/python3.10/site-packages/pystarlark/starlark_go.cpython-310-x86_64-linux-gnu.so"
rm "$PYSTARLARK_SO"
env CGO_CFLAGS="-g -O0 $(python3-config --includes)" \
  CGO_LDFLAGS=-Wl,--unresolved-symbols=ignore-all \
  go build -buildmode=c-shared -o "$PYSTARLARK_SO"

# Remove old log and then run valgrind
rm -f "$VALGRIND_LOG"
if ! valgrind --gen-suppressions=all --suppressions=scripts/pytest-valgrind.supp \
  --show-leak-kinds=definite --errors-for-leak-kinds=definite --log-file="$VALGRIND_LOG" \
  "$VALGRIND_PYTHON" -m pytest -vv --valgrind --valgrind-log="$VALGRIND_LOG"; then
  set +x
  echo
  echo "*** VALGRIND FAILED, FULL LOG FOLLOWS ***"
  echo
  cat "$VALGRIND_LOG"
  exit 1
fi
