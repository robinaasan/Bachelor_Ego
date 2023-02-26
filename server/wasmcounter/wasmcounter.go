package wasmcounter

import (
	"github.com/wasmerio/wasmer-go/wasmer"
)

type MyEnvironment struct {
	shift int32
}

// function that takes as paramters: *wasmer.Engine, *environment, []byte with wasm module,
// return the instance
func GetNewWasmInstace(engine *wasmer.Engine, store *wasmer.Store, n int32, i []byte) (*wasmer.Instance, error) {
	// Create a new module from some WebAssembly in its text representation
	// (for the sake of simplicity of the example).
	//engine := wasmer.NewEngine()

	// Create a store, that holds the engine.

	environment := &MyEnvironment{
		shift: n,
	}

	module, _ := wasmer.NewModule(
		store,
		[]byte(`
			(module
			;; We import a math.sum function.
			(import "math" "sum" (func $sum (param i32 i32) (result i32)))

			;; We export an add_one function.
			(func (export "add_one") (param $x i32) (result i32)
				local.get $x
				i32.const 1
				call $sum))
		`),
	)

	// Let's create a new host function for `math.sum`.
	function := wasmer.NewFunctionWithEnvironment(
		store,

		// The function signature.
		wasmer.NewFunctionType(
			// Parameters.
			wasmer.NewValueTypes(wasmer.I32, wasmer.I32),
			// Results.
			wasmer.NewValueTypes(wasmer.I32),
		),
		environment,

		// The function implementation.

		//TODO: change name of environment?
		func(environment interface{}, args []wasmer.Value) ([]wasmer.Value, error) {
			// Cast to our environment type, and do whatever we want!
			env := environment.(*MyEnvironment)
			e := env.shift
			x := args[0].I32()
			y := args[1].I32()

			return []wasmer.Value{wasmer.NewI32(e + x + y)}, nil
		},
	)

	// Let's use the new `ImportObject` API…
	importObject := wasmer.NewImportObject()

	//… to register the `math.sum` function.
	importObject.Register(
		"math",
		map[string]wasmer.IntoExtern{
			"sum": function,
		},
	)

	// Finally, let's instantiate the module, with all the imports.
	instance, err := wasmer.NewInstance(module, importObject)

	if err != nil {
		return nil, err
	}
	// And let's call the `add_one` function!
	//addOne, _ := instance.Exports.GetFunction("add_one")

	return instance, nil
}

// func main() {
// 	wasmBytes, err := ioutil.ReadFile("rust_host_func.wasm")

// 	if err != nil {
// 		fmt.Println("Error:", err)
// 	}

// 	environment := &MyEnvironment{
// 		foo: 42,
// 	}
// 	engine := wasmer.NewEngine()
// 	store := wasmer.NewStore(engine)

// 	module, err := wasmer.NewModule(store, wasmBytes)

// 	if err != nil {
// 		fmt.Println("ien")
// 	}

// 	importObject := wasmer.NewImportObject()
// 	instance, err := wasmer.NewInstance(module, importObject)

// 	run, err := instance.Exports.GetFunction("run")

// 	if err != nil {
// 		fmt.Println("Error!:", err)
// 	}
// 	hostFunction := wasmer.NewFunctionWithEnvironment(
// 		store,
// 		wasmer.NewFunctionType(
// 			wasmer.NewValueTypes(),
// 			wasmer.NewValueTypes(wasmer.I32),
// 		),
// 		environment,
// 		func(environment interface{}, args []wasmer.Value) ([]wasmer.Value, error) {
// 			en := environment.(*MyEnvironment)
// 			return []wasmer.Value{wasmer.NewI32(42)}, nil
// 		},
// 	)
// 	res, err := run()
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	fmt.Println(res)

// }

// func (e *MyEnvironment) add(a int32, b int32) int32 {
// 	return a + b
// }
