from typing import Any, Callable, List, Mapping, Optional

def configure_starlark(
    *,
    allow_set: Optional[bool] = ...,
    allow_global_reassign: Optional[bool] = ...,
    allow_recursion: Optional[bool] = ...,
) -> None: ...

class Starlark:
    def __init__(
        self,
        *,
        globals: Optional[Mapping[str, Any]] = ...,
        print: Callable[[str], Any] = ...,
    ) -> None: ...
    def eval(
        self,
        expr: str,
        *,
        filename: Optional[str] = ...,
        convert: Optional[bool] = ...,
        print: Callable[[str], Any] = ...,
    ) -> Any: ...
    def exec(
        self,
        defs: str,
        *,
        filename: Optional[str] = ...,
        print: Callable[[str], Any] = ...,
    ) -> None: ...
    def globals(self) -> List[str]: ...
    def get(self, name: str, default_value: Optional[Any] = ...) -> None: ...
    def set(self, **kwargs: Any) -> None: ...
    def pop(self, name: str, default_value: Optional[Any] = ...) -> Any: ...
    @property
    def print(self) -> Optional[Callable[[str], Any]]: ...
    @print.setter
    def print(
        self, value: Optional[Callable[[str], Any]]
    ) -> Optional[Callable[[str], Any]]: ...
