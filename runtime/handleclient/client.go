package handleclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

// structure for each vendor with their respective name (Hash), 
// uploaded wasm file and the created wasm instance and wasm function
type Client struct {
	Hash           []byte
	Wasm_file      *WasmFile
	Wasm           *WasmerGO
	ClientMessages map[string]bool // map for a unique message Id (string) set to true if the client (vendor) sent a message 
	WaitForAckFromOrdering chan string // channel for waiting for ack after client sending the message (transaction) to ordering
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
		ClientMessages: make(map[string]bool),
		WaitForAckFromOrdering: make(chan string),
	}
}

// function for using the wasm function, return the retrieved values
func (cl *Client) UseWasmFunction(key int, value int, runtime *Runtime) (*SetValue, error) {
	setvalues := SetValue{0, 0, 0}

	// various commands for the wasm function
	// fmt.Println(cl.Wasm.Function.Type())
	// fmt.Println(cl.Wasm.Function.ParameterArity())
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

// set the wasm module for the client
func (cl *Client) SetWasmFile(r *http.Request) error {
	wasmfile := cl.Wasm_file
	err := json.NewDecoder(r.Body).Decode(wasmfile)
	if err != nil {
		return err
	}
	return nil
}

// Confirm that the client has uploaded a wasm file
func (cl *Client) WasmFileExist() bool {
	return len(cl.Wasm_file.File) != 0
}

func GetClient(hash string, allClients AllClients) (*Client) {
	cl, exists := allClients[hash]
	if exists {
		return cl
	}
	return &Client{}
}
