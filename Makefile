build:
	go build -buildmode=c-shared -o starlark.so .
	python build_ffi.py

clean:
	rm starlark.c
	rm starlark.h
	rm starlark.o
	rm starlark.so