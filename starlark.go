package main

import "C"
import "go.starlark.net/starlark"
import "fmt"
import "math"

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

//export Hello
func Hello() {
	fmt.Printf("Hello! The square root of 4 is: %g\n", math.Sqrt(4))
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
