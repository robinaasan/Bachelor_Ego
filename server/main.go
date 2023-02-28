package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/robinaasan/Bachelor_Ego/server/wasmcounter"
	wasmer "github.com/wasmerio/wasmer-go/wasmer"
)

type WasmFile struct {
	File []byte
}

func newWasmFile() *WasmFile {
	return &WasmFile{
		File: []byte{},
	}
}

var wasm_file = newWasmFile()
var env = wasmcounter.MyEnvironment{Shift: int32(0)}

//var storage = newStorage()

func handlerAdd(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])

	if len(wasm_file.File) == 0 {
		fmt.Fprintf(w, "There is no wasm file here!")
		return
	}

	query := r.URL.Query()
	fmt.Println(query)

	var query_key_val1 int

	query_key_val1, err := strconv.Atoi(query.Get("val1"))

	if err != nil {
		//he is probably uploading a file
		fmt.Fprint(w, err)
	}
	err = useWasmFunction(wasm_file, query_key_val1)

	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w, err)
	}

}

func getWasmFile(r *http.Request) error {

	//r.ParseMultipartForm(32 << 20) // limit your max input length!
	//var buf bytes.Buffer
	// in your case file would be fileupload
	//file, header, err := r.FormFile("file")

	err := json.NewDecoder(r.Body).Decode(&wasm_file)

	if err != nil {
		return err
	}

	fmt.Printf("Json: %v", string(wasm_file.File))

	// fmt.Printf("Filename: %v\n", name_type[0])
	// io.Copy(&buf, file)

	return nil
}

func useWasmFunction(wasm_file *WasmFile, value1 int) error {

	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	// Compiles the module
	instance, err := wasmcounter.GetNewWasmInstace(&env, engine, store, wasm_file.File)

	if err != nil {
		return err
	}

	addOne, err := instance.Exports.GetRawFunction("add_one")

	if err != nil {
		return err
	}
	fmt.Println(addOne.Type())
	//fmt.Println(addOne.ParameterArity())
	//fmt.Println(addOne.ResultArity())
	result, err := addOne.Call(value1)

	if err != nil {
		return err
	}

	fmt.Printf("Returned: %v, Shift value: %v\n", result, env.Shift)

	return nil
}

func handlerUpload(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	fmt.Println(query)

	err := getWasmFile(r)

	if err != nil {
		fmt.Fprint(w, err)
	}
}

func main() {

	http.HandleFunc("/Add", handlerAdd)
	http.HandleFunc("/Upload", handlerUpload)

	//tlsConfig, err := enclave.CreateAttestationServerTLSConfig()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	server := http.Server{Addr: ":8081"}
	fmt.Println("Listening...")
	err := server.ListenAndServe()
	// server.ListenAndServeTLS("", "")
	fmt.Println(err)

	// debug.PrintStack()

}
