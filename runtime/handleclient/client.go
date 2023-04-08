package handleclient

import (
	"errors"
	"fmt"

	"github.com/benpate/convert"
)

type WasmFile struct {
	File []byte `json:"File"`
}

func NewWasmFile() *WasmFile {
	return &WasmFile{
		File: []byte{},
	}
}

// structure for each vendor with their respective name (Hasb), uploaded wasmfile and the created wasm instane and wasm function
type Client struct {
	Hash      []byte
	Wasm_file *WasmFile
	Wasm      *WasmerGO
}

type SetValue struct {
	Key    int
	NewVal int
	OldVal int
}

func NewClient(name string) *Client {
	return &Client{
		Hash:      []byte(name),  // name of the client
		Wasm_file: NewWasmFile(), // Wasm_file is the file in bytes
		Wasm:      NewWasmerGO(), // Wasm contains the wasm function and instace
	}
}

type AllClients map[string]*Client

// function for using the wasm function, return the retrieved values
func (cl *Client) UseWasmFunction(key int, value int, runtime *Runtime) (*SetValue, error) {
	setvalues := SetValue{0, 0, 0}

	// fmt.Println(cl.Wasm.Function.Type())
	//fmt.Println(cl.Wasm.Function.ParameterArity())
	// fmt.Println(cl.Wasm.Function.ResultArity())

	result, err := cl.Wasm.Function.Call(key, value)
	if err != nil {
		return &setvalues, err
	}

	// convert the returned interface to a slice of ints
	nl := convert.SliceOfInt(result)
	setvalues.Key, setvalues.NewVal, setvalues.OldVal = nl[0], nl[1], nl[2]
	if err != nil {
		return &setvalues, err
	}
	return &setvalues, nil
}

// create the instance for the vendor
func (cl *Client) CreateInstanceClient(runtime *Runtime) error {
	// check if the instance already exists

	fmt.Println("Creating Instance...")
	var err error
	cl.Wasm.Instance, err = runtime.GetNewWasmInstace(cl.Wasm_file.File)
	if err != nil {
		return err
	}
	smart_contract, err := cl.Wasm.Instance.Exports.GetRawFunction("add_one")
	if err != nil {
		return err
	}
	cl.Wasm.Function = smart_contract

	if cl.Wasm.Function == nil {
		return errors.New("error: the function for the client isn't set")
	}
	return nil
}
