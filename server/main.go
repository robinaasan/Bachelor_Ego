package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/edgelesssys/ego/ecrypto"
	"github.com/edgelesssys/ego/enclave"

	//"github.com/edgelesssys/ego/enclave"
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
var env = wasmcounter.NewEnvironment()
var wasmer_module = wasmcounter.WasmerGO{Instance: nil, Function: nil}

//var storage = newStorage()

func handlerSet(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])

	// serverURL := flag.String("url", "localhost:8083", "Server's url")
	// flag.Parse()

	// req := url.URL{Scheme: "http", Host: *serverURL, Path: "/"}
	// q := url.Values{}

	// req.RawQuery = q.Encode()
	// client := http.Client{}
	// resp, err := client.Get(req.String())
	// if err != nil {
	// 	fmt.Fprint(w, err)
	// }
	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Fprint(w, err)
	// }
	// fmt.Println(string(body))

	//Slutt client ordering
	if len(wasm_file.File) == 0 {
		fmt.Fprintf(w, "There is no wasm file here!")
		return
	}

	query := r.URL.Query()
	fmt.Println(query)

	key, err := strconv.Atoi(query.Get("key"))
	value, err := strconv.Atoi(query.Get("value"))

	if err != nil {
		//he is probably uploading a file
		fmt.Fprint(w, err)
	}
	err = useWasmFunction(wasm_file, key, value)

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

func useWasmFunction(wasm_file *WasmFile, key int, value int) error {
	//TODO: check if this need to be loaded for each SET operation
	err := LoadState()
	if err != nil {
		return err
	}
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	//check if the instance already exists
	if wasmer_module.Instance == nil {
		fmt.Println("Creating Instance...")
		instance, err := wasmcounter.GetNewWasmInstace(env, engine, store, wasm_file.File)
		if err != nil {
			return err
		}
		wasmer_module.Instance = instance
		addOne, err := wasmer_module.Instance.Exports.GetRawFunction("add_one")

		if err != nil {
			return err
		}
		wasmer_module.Function = addOne
	}

	fmt.Println(wasmer_module.Function.Type())
	//fmt.Println(addOne.ParameterArity())
	//fmt.Println(addOne.ResultArity())
	result, err := wasmer_module.Function.Call(key, value)

	if err != nil {
		return err
	}

	//Write to the store!
	mustSaveState()
	fmt.Printf("Returned: %v, Store value: %v\n", result, env.Store)

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
	// err := mustSaveState()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	http.HandleFunc("/Add", handlerSet)
	http.HandleFunc("/Upload", handlerUpload)

	//embeds certificate on its own by default
	tlsConfig, err := enclave.CreateAttestationServerTLSConfig()
	if err != nil {
		log.Fatal(err)
	}
	server := http.Server{Addr: ":8082", TLSConfig: tlsConfig}
	fmt.Println("Listening...")
	//err := server.ListenAndServe()
	err = server.ListenAndServeTLS("", "")

	//err := LoadState()
	if err != nil {
		fmt.Println("Error here!", err)
		return
	}
	//fmt.Printf("%v\n", env.Store)
	// debug.PrintStack()

}

// Stores secrets to disk
//TODO: All below is partly copied code!
func mustSaveState() error {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)

	// Encoding state
	err := e.Encode(env.Store)
	if err != nil {
		return err
	}
	encState, err := ecrypto.SealWithProductKey(b.Bytes(), nil)
	if err != nil {
		return err
	}

	if err := os.WriteFile("/data/secret.store", encState, 0600); err != nil {
		return fmt.Errorf("Error: creating file responded with: %v", err)
	}
	return nil
}

//read the file and set map in env from storage
func LoadState() error {
	file, err := os.ReadFile("/data/secret.store")
	if os.IsNotExist(err) {

		fmt.Println("The file does not exist, creating one in this enclave ...")
		mustSaveState()
	}
	//the storage exists
	decrypted_file, err := ecrypto.Unseal(file, nil)
	if err != nil {
		fmt.Println("Error unsealing...")
		return err
	}

	dec := gob.NewDecoder(bytes.NewBuffer(decrypted_file))
	err = dec.Decode(&env.Store)
	if err != nil {
		return err
	}
	return nil
}
