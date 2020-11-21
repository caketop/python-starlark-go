import pystarlark

fib = b"""
def fibonacci(n=10):
   res = list(range(n))
   for i in res[2:]:
       res[i] = res[i-2] + res[i-1]
   return res
"""

fib2 = b"""
'test'.capitalize()
"""

# response = lib.ExecCall(fib, b"fibonacci")
# output = ffi.string(response)
# print(output)
# print(json.loads(output))

# response = lib.ExecEval(fib2)
# output = ffi.string(response)
# print(output)
# print(json.loads(output))

response = pystarlark.ExecCallEval(fib, "fibonacci(n=20)")
print(response)
