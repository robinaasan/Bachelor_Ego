package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	wasmer "github.com/wasmerio/wasmer-go/wasmer"
)

type WasmFile struct {
	fileBytes []byte
}

type Storage struct {
	number int32
}

func newWasmFile() *WasmFile {
	return &WasmFile{
		fileBytes: []byte{},
	}
}

func newStorage() *Storage {
	return &Storage{
		number: 0,
	}
}

var wasm_file = newWasmFile()
var storage = newStorage()

func handlerAdd(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])

	if len(wasm_file.fileBytes) == 0 {
		fmt.Fprintf(w, "There is no wasm file here!")
		return
	}
	query := r.URL.Query()
	fmt.Println(query)
	//cmd := query.Get("cmd")
	var query_key_val1, query_key_val2 int

	query_key_val1, err := strconv.Atoi(query.Get("val1"))

	if err != nil {
		//he is probably uploading a file
		//fmt.Fprint(w, err)
	}
	query_key_val2, err = strconv.Atoi(query.Get("val2"))
	if err != nil {
		fmt.Fprint(w, err)
	}
	err = useWasmFunction(wasm_file, query_key_val1, query_key_val2)

	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w, err)
	}
	// value1, err := strconv.Atoi(query.Get("val1"))
	// //value2 := query.Get("val2")
	// value2, err := strconv.Atoi(query.Get("val2"))

	/* 	fmt.Println(cmd)
	   	fmt.Println(value1)
	   	fmt.Println(value2)
	*/
	//contents := buf.String()
	//fmt.Println(contents)

}

func getWasmFile(r *http.Request) error {

	r.ParseMultipartForm(32 << 20) // limit your max input length!
	var buf bytes.Buffer
	// in your case file would be fileupload
	file, header, err := r.FormFile("file")

	if err != nil {
		return fmt.Errorf(err.Error())

		//panic(err)
	}

	defer file.Close()
	name_type := strings.Split(header.Filename, ".")

	if name_type[1] == "wasm" {
		fmt.Println("Correct file type!")
	} else {
		return fmt.Errorf("Wanted file type: txt and got filetype: %v", name_type[1])

		//os.Exit(1)
	}

	fmt.Printf("Filename: %v\n", name_type[0])
	io.Copy(&buf, file)

	//	wasmBytes :=

	wasm_file.fileBytes = buf.Bytes()

	return nil

	// if err != nil {
	// 	fmt.Errorf(err.Error())
	// }
}

func useWasmFunction(wasm_file *WasmFile, value1 int, value2 int) error {
	fmt.Println("Function runs:)")
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	// Compiles the module
	module, err := wasmer.NewModule(store, wasm_file.fileBytes)

	if err != nil {
		return fmt.Errorf("Failed to compile module Error:\n%v", err)
	}
	// Instantiates the module
	importObject := wasmer.NewImportObject()
	instance, err := wasmer.NewInstance(module, importObject)

	if err != nil {
		return fmt.Errorf("Failed to instaciate module Error:\n%v", err)
	}
	// Gets the `sum` exported function from the WebAssembly instance.
	sum, err := instance.Exports.GetFunction("sum")

	if err != nil {
		return fmt.Errorf("Failed to get function from module Error:\n%v", err)
	}
	// Calls that exported function with Go standard values. The WebAssembly
	// types are inferred and values are casted automatically.
	result, _ := sum(value1, value2)

	storage.number += result.(int32)
	//storage.number = storage.number + result
	fmt.Println("running storage..")
	fmt.Println(storage.number) // 42!
	return nil
}

func handlerUpload(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	fmt.Println(query)

	//_, err := strconv.Atoi(query.Get("filename"))
	// if err != nil {
	// 	fmt.Fprint(w, err)
	// }

	err := getWasmFile(r)

	if err != nil {
		fmt.Fprint(w, err)
	}
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/Add", handlerAdd)
	mux.HandleFunc("/Upload", handlerUpload)

	//tlsConfig, err := enclave.CreateAttestationServerTLSConfig()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//	server := http.Server{Addr: ":8080", TLSConfig: tlsConfig}
	server := http.Server{Addr: ":8080"}

	server.ListenAndServeTLS("", "")
	//http.ListenAndServe(":8080", nil)

}
