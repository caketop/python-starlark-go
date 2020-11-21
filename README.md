go build -buildmode=c-shared -o starlark.so .

python build_ffi.py