from pystarlark import Starlark

s = Starlark()
fibonacci = """
def fibonacci(n=10):
   res = list(range(n))
   for i in res[2:]:
       res[i] = res[i-2] + res[i-1]
   return res
"""
s.read(fibonacci)
s.eval("fibonacci(5)")  # [0, 1, 1, 2, 3]