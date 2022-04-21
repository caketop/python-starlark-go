from typing import Any, List, Mapping, Optional

def configure_starlark(
    *,
    allow_set: Optional[bool] = ...,
    allow_global_reassign: Optional[bool] = ...,
    allow_recursion: Optional[bool] = ...,
) -> None: ...

class Starlark:
    def __init__(self, *, globals: Optional[Mapping[str, Any]] = ...) -> None: ...
    def eval(
        self, expr: str, *, filename: Optional[str] = ..., convert: Optional[bool] = ...
    ) -> Any: ...
    def exec(self, defs: str, *, filename: Optional[str] = ...) -> None: ...
    def globals(self) -> List[str]: ...
    def get(self, name: str, default_value: Optional[Any] = ...) -> None: ...
    def set(self, **kwargs: Any) -> None: ...
    def pop(self, name: str, default_value: Optional[Any] = ...) -> Any: ...
