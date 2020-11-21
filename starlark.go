package main

import (
	"C"
	"encoding/json"
	"fmt"
	"go.starlark.net/starlark"
)

func NewStarlark() *starlark.Thread {
	thread := &starlark.Thread{}
	return thread
}

func ExecFile(thread *starlark.Thread, data string) (starlark.StringDict, error) {
	return starlark.ExecFile(thread, "main.star", data, starlark.StringDict{})
}

func Call(thread *starlark.Thread, globals starlark.StringDict, fname string, args starlark.Tuple, kwargs []starlark.Tuple) string {
	// Run star from a string

	f := globals[fname]

	// // Call Starlark function from Go.
	// new_kwargs := starlark.StringDict{
	// 	"n": starlark.MakeInt(25),
	// }

	// make(map[string]starlark.Value)}

	v, _ := starlark.Call(thread, f, args, kwargs)
	return v.String()
}

//export ExecCall
func ExecCall(data *C.char, function *C.char) *C.char {
	goData := C.GoString(data)
	goFunction := C.GoString(function)
	thread := NewStarlark()
	globals, _ := ExecFile(thread, goData)

	result := Call(thread, globals, goFunction, nil, nil)

	return C.CString(result)
}

//export ExecCallEval
func ExecCallEval(preamble *C.char, statement *C.char) *C.char {
	// Cast *char to string
	goPreamble := C.GoString(preamble)
	goStatement := C.GoString(statement)

	// Initialize starlark and execute preamble
	thread := NewStarlark()
	globals, _ := ExecFile(thread, goPreamble)

	// Execute statement
	result, _ := starlark.Eval(thread, "<expr>", goStatement, globals)

	// Convert starlark.Value struct into a JSON blob
	rawResponse := make(map[string]string)
	rawResponse["value"] = result.String()
	rawResponse["type"] = result.Type()
	response, _ := json.Marshal(rawResponse)

	// Convert JSON blob to string and then CString
	return C.CString(string(response))
}

//export ExecEval
func ExecEval(data *C.char) *C.char {
	stmt := C.GoString(data)
	thread := NewStarlark()

	result, err := starlark.Eval(thread, "<expr>", stmt, nil)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return C.CString(result.String())
}

//export ExecCall
// func ExecCall(data string, fname string, kwargs map[string]interface{}) string {
// 	thread := NewStarlark()
// 	globals, _ := ExecFile(thread, data)
// 	// fmt.Printf("%v\n", kwargs)

// 	// new_kwargs := []starlark.Tuple{
// 	// 	starlark.MakeInt(6),
// 	// }

// 	// args := starlark.Tuple{starlark.MakeInt(6)}
// 	// new_kwargs := []starlark.Tuple{
// 	// 	starlark.Tuple{starlark.String("n"), starlark.MakeInt(20)},
// 	// }

// 	return Call(thread, globals, fname, nil, nil)
// }

func Eval(thread *starlark.Thread, globals starlark.StringDict, stmt string) string {
	v, err := starlark.Eval(thread, "<expr>", stmt, globals)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return v.String()
}

func main() {
	const data = `
def fibonacci(n=10):
	res = list(range(n))
	for i in res[2:]:
		res[i] = res[i-2] + res[i-1]
	return res
`
	r := ExecCallEval(C.CString(data), C.CString("fibonacci()"))
	fmt.Printf("%v\n", C.GoString(r))

	// kwargs := map[string]interface{}{
	// 	"n": 4,
	// }
	// r := ExecCall(data, "fibonacci", nil)
	// fmt.Printf("%v\n", r)

	// thread := NewStarlark()
	// fmt.Println(reflect.TypeOf(thread))
	// globals, _ := ExecFile(thread, data)
	// fmt.Printf("%v\n", globals)
	// Call(thread, globals, "fibonacci", starlark.Tuple{starlark.MakeInt(10)}, nil)

	// Eval(thread, globals, "None")
	// _, prog, _ := Build(data)
	// globals, _ := prog.Init(thread, starlark.StringDict{})
	// fmt.Printf("%v\n", globals)
	// prog.Call
}
