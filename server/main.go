package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	wasmer "github.com/wasmerio/wasmer-go/wasmer"
)

func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])

	query := r.URL.Query()
	fmt.Println(query)
	//cmd := query.Get("cmd")
	value1, err := strconv.Atoi(query.Get("val1"))
	//value2 := query.Get("val2")
	value2, err := strconv.Atoi(query.Get("val2"))

	/* 	fmt.Println(cmd)
	   	fmt.Println(value1)
	   	fmt.Println(value2)
	*/
	r.ParseMultipartForm(32 << 20) // limit your max input length!
	var buf bytes.Buffer
	// in your case file would be fileupload
	file, header, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	name_type := strings.Split(header.Filename, ".")

	if name_type[1] == "wasm" {
		fmt.Println("Correct file type!")
	} else {
		fmt.Fprintf(w, "Wanted file type: txt and got filetype: %v", name_type[1])
		os.Exit(1)
	}

	fmt.Printf("Filename: %v\n", name_type[0])
	io.Copy(&buf, file)

	wasmBytes := buf.Bytes()

	err = useWasmFunction(wasmBytes, value1, value2)

	if err != nil {
		fmt.Fprint(w, err)
	}
	//contents := buf.String()
	//fmt.Println(contents)
}

func useWasmFunction(wasmBytes []byte, value1 int, value2 int) error {
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	// Compiles the module
	module, err := wasmer.NewModule(store, wasmBytes)

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

	fmt.Println(result) // 42!
	return nil
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
