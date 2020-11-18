from pystarlark import lib, ffi
import json

fib = b"""
def fibonacci(n=10):
   res = list(range(n))
   for i in res[2:]:
       res[i] = res[i-2] + res[i-1]
   return res
"""

response = lib.ExecCall(fib, b"fibonacci")
output = ffi.string(response)

print(json.loads(output))
