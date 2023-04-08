package handleclient

import (
	wasmer "github.com/wasmerio/wasmer-go/wasmer"
)

// struct for holding the wasmer instance and function
type WasmerGO struct {
	Instance *wasmer.Instance
	Function *wasmer.Function
}

func NewWasmerGO() *WasmerGO {
	return &WasmerGO{}
}

// Create a new wasmer instance for a client. 
//Wasmmodule given as param.
//Link: https://wasmer.io/posts/wasmer-go-embedding-1.0
func (runtime *Runtime) GetNewWasmInstace(fileBytes []byte) (*wasmer.Instance, error) {
	// Create a new module from some WebAssembly in its text representation
	// (for the sake of simplicity of the example).
	// Create a store, that holds the engine.
	module, _ := wasmer.NewModule(
		runtime.WasmStore,
		fileBytes,
	)
	// Create a new host function for `math.set`.
	function := wasmer.NewFunctionWithEnvironment(
		runtime.WasmStore,
		// The function signature.
		wasmer.NewFunctionType(
			// Parameters.
			wasmer.NewValueTypes(wasmer.I32, wasmer.I32),
			// Results.
			wasmer.NewValueTypes(wasmer.I32, wasmer.I32, wasmer.I32),
		),
		runtime.Environment,
		// The function implementation.
		func(environment interface{}, args []wasmer.Value) ([]wasmer.Value, error) {
			// Cast to our environment type
			env := environment.(*EnvStore)
			x := args[0].I32() //key
			y := args[1].I32() + 1//val
			oldVal, exists := env.Store[x]
			// (*env).Store[x] = y
			if exists {
				return []wasmer.Value{wasmer.NewI32(x), wasmer.NewI32(y), wasmer.NewI32(oldVal)}, nil
			}
			return []wasmer.Value{wasmer.NewI32(x), wasmer.NewI32(y), wasmer.NewI32(0)}, nil
		},
	)

	// use the new `ImportObject` API…
	importObject := wasmer.NewImportObject()
	//… to register the `math.set` function.
	importObject.Register(
		"math",
		map[string]wasmer.IntoExtern{
			"set": function,
		},
	)
	//Instantiate the module, with all the imports.
	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		return nil, err
	}
	return instance, nil
}
