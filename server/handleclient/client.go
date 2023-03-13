package handleclient

import (
	"errors"
	"fmt"

	"github.com/benpate/convert"
	"github.com/robinaasan/Bachelor_Ego/server/wasmcounter"
	"github.com/wasmerio/wasmer-go/wasmer"
)

type WasmFile struct {
	File []byte `json:"File"`
}

func NewWasmFile() *WasmFile {
	return &WasmFile{
		File: []byte{},
	}
}

type Client struct {
	Hash      []byte
	Wasm_file *WasmFile
	Wasm      *wasmcounter.WasmerGO
}

//the string is hashed and sent to client so the server know which client to use
//var AllClients map[[32]byte]*Client
var AllClients = make(map[string]*Client)

type SetValue struct {
	Key    int
	NewVal int
	OldVal int
}

//The hash should probably be created beforehand. For now the clients are seperated by their name
//Wasm_file is the file in bytes
//Wasm contains the wasm function and instace
func NewClient(name string) *Client {
	// h := sha256.New()
	// h.Write([]byte(name))
	return &Client{
		Hash:      []byte(name),
		Wasm_file: NewWasmFile(),
		Wasm:      wasmcounter.NewWasmerGO(),
	}
}

func (cl *Client) UseWasmFunction(key int, value int, env *wasmcounter.MyEnvironment, engine *wasmer.Engine, store *wasmer.Store) (SetValue, error) {
	setvalues := SetValue{0, 0, 0}
	//check if the instance already exists
	if cl.Wasm.Instance == nil {
		fmt.Println("Creating Instance...")
		var err error
		cl.Wasm.Instance, err = wasmcounter.GetNewWasmInstace(env, engine, store, cl.Wasm_file.File) //See global variabled
		if err != nil {
			return setvalues, err
		}
		//TODO: change name from add_one
		smart_contract, err := cl.Wasm.Instance.Exports.GetRawFunction("add_one")
		if err != nil {
			return setvalues, err
		}
		cl.Wasm.Function = smart_contract
	}
	if cl.Wasm.Function == nil {
		return setvalues, errors.New("error: the function for the client isn't set")
	}

	fmt.Println(cl.Wasm.Function.Type())
	//fmt.Println(smart_contract.ParameterArity())
	//fmt.Println(smart_contract.ResultArity())
	result, err := cl.Wasm.Function.Call(key, value)
	if err != nil {
		return setvalues, err
	}
	nl := convert.SliceOfInt(result)
	key, newVal, oldVal := nl[0], nl[1], nl[2]
	setvalues.Key = key
	setvalues.NewVal = newVal
	setvalues.OldVal = oldVal
	//TODO: Send to ordering service store!
	if err != nil {
		return setvalues, err
	}
	//TODO: write function that notices about a change to orderingservice
	//fmt.Printf("key= %v V=%v N=%v", key, newVal, oldVal)
	//fmt.Printf("Returned: %v, Type:%T Store value: %v\n", nl, nl, env.Store)
	return setvalues, nil
}
