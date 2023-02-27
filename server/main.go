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

	// if len(wasm_file.fileBytes) == 0 {
	// 	fmt.Fprintf(w, "There is no wasm file here!")
	// 	return
	// }

	query := r.URL.Query()

	//command := query.Get("cmd")
	// if command == "add" {
	// 	//Add the numbers
	// } else if command == "upload" {
	// 	//upload the stuff
	// } else {
	// 	fmt.Fprint(w, errors.New("Error: No parameter included"))

	// }
	fmt.Println(query)
	//cmd := query.Get("cmd")
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
	// decoder := json.NewDecoder(r.Body)

	// err := decoder.Decode(&wasm_file)
	//body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	return err
	// }

	err := json.NewDecoder(r.Body).Decode(&wasm_file)

	if err != nil {
		return err
	}

	fmt.Printf("Json: %v", string(wasm_file.File))

	// if err != nil {

	// 	return fmt.Errorf(err.Error())

	// 	//panic(err)
	// }

	//defer file.Close()
	// name_type := strings.Split(header.Filename, ".")

	// if name_type[1] == "wasm" {
	// 	fmt.Println("Correct file type!")
	// } else {
	// 	return fmt.Errorf("Wanted file type: txt and got filetype: %v", name_type[1])

	// 	//os.Exit(1)
	// }

	// fmt.Printf("Filename: %v\n", name_type[0])
	// io.Copy(&buf, file)

	//	wasmBytes :=

	//wasm_file.fileBytes = buf.Bytes()

	return nil

	// if err != nil {
	// 	fmt.Errorf(err.Error())
	// }
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

// func handlerUpload(w http.ResponseWriter, r *http.Request) {
// 	query := r.URL.Query()
// 	fmt.Println(query)

// 	//_, err := strconv.Atoi(query.Get("filename"))
// 	// if err != nil {
// 	// 	fmt.Fprint(w, err)
// 	// }

// 	err := getWasmFile(r)

// 	if err != nil {
// 		fmt.Fprint(w, err)
// 	}
// }

func main() {

	//mux := http.NewServeMux()

	wasm_file.File = []byte(`
	(module
		;; We import a math.sum function.
		(import "math" "sum" (func $sum (param i32 i32) (result i32)))

		;; We export an add_one function.
		(func (export "add_one") (param $x i32) (result i32)
			local.get $x
			i32.const 1
			call $sum))
	`)

	http.HandleFunc("/Add", handlerAdd)
	//.HandleFunc("/Upload", handlerUpload)

	//tlsConfig, err := enclave.CreateAttestationServerTLSConfig()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	server := http.Server{Addr: ":8080"}
	fmt.Println("Listening...")
	err := server.ListenAndServe()

	fmt.Println(err)
	//http.ListenAndServe(":8080", nil)

	// debug.PrintStack()

	// server.ListenAndServeTLS("", "")
	//server.ListenAndServe()

	//http.ListenAndServe(":8080", nil)

}
