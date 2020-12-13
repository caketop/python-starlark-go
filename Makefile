so:
	go build -buildmode=c-shared -o starlark.so .

ffi:
	python build_ffi.py

clean:
	rm -rf starlark.c
	rm -rf starlark.h
	rm -rf starlark.o
	rm -rf starlark.so

test:
	python -m pytest