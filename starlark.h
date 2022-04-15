typedef struct StarlarkErrorArgs {
  char *error;
  char *error_type;
} StarlarkErrorArgs;

typedef struct SyntaxErrorArgs {
  char *error;
  char *error_type;
  char *msg;
  char *filename;
  unsigned int line;
  unsigned int column;
} SyntaxErrorArgs;

typedef struct EvalErrorArgs {
  char *error;
  char *error_type;
  char *backtrace;
} EvalErrorArgs;

typedef enum StarlarkErrorType {
  STARLARK_NO_ERROR = 0,
  STARLARK_GENERAL_ERROR = 1,
  STARLARK_SYNTAX_ERROR = 2,
  STARLARK_EVAL_ERROR = 3
} StarlarkErrorType;

typedef struct StarlarkReturn {
  char *value;
  StarlarkErrorType error_type;
  void *error;
} StarlarkReturn;
