#!/usr/bin/env bash

set -euo pipefail

OLD_STARLARK_VERSION=$(grep 'require go.starlark.net' go.mod | cut -d' ' -f3)

go get -u go.starlark.net

NEW_STARLARK_VERSION=$(grep 'require go.starlark.net' go.mod | cut -d' ' -f3)

if [ "$NEW_STARLARK_VERSION" = "$OLD_STARLARK_VERSION" ]; then
  echo "starlark-go is unchanged (still $NEW_STARLARK_VERSION)"
  exit 0
fi

# shellcheck disable=SC2260
if sed --version 2&>/dev/null | grep -q GNU ; then
  sed -i "s/$OLD_STARLARK_VERSION/$NEW_STARLARK_VERSION/g" README.md
else
  sed -i .bak "s/$OLD_STARLARK_VERSION/$NEW_STARLARK_VERSION/g" README.md
  rm -f README.md.bak || true
fi

if [ -n "$GITHUB_ENV" ]; then
  echo "NEW_STARLARK_VERSION=$NEW_STARLARK_VERSION" >> "$GITHUB_ENV"
fi
