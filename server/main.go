package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/edgelesssys/ego/ecrypto"

	//"github.com/edgelesssys/ego/enclave"
	"github.com/benpate/convert"
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

// type ClientsList struct {

// }

const orderingURL = "http://localhost:8087"

var wasm_file = newWasmFile()
var env = wasmcounter.NewEnvironment()
var engine = wasmer.NewEngine()
var store = wasmer.NewStore(engine)
var wasmer_module = wasmcounter.WasmerGO{Instance: nil, Function: nil}

//var storage = newStorage()

func handlerSet(w http.ResponseWriter, r *http.Request) {
	if len(wasm_file.File) == 0 {
		fmt.Fprintf(w, "There is no wasm file here!")
		return
	}
	query := r.URL.Query()
	fmt.Println(query)
	var key, value int

	key, err := strconv.Atoi(query.Get("key"))
	if err != nil {
		fmt.Fprintf(w, "Error: couldn't get the key")
	} else {
		value, err = strconv.Atoi(query.Get("value"))
		if err != nil {
			fmt.Fprintf(w, "Error: couldn't get the value")
		}
	}
	//For each request use the wasmFunction to get the wasm module
	err = useWasmFunction(wasm_file, key, value)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintln(w, err)
	}
}

func getWasmFile(r *http.Request) error {
	err := json.NewDecoder(r.Body).Decode(&wasm_file)
	if err != nil {
		return err
	}
	fmt.Printf("Json: %v", string(wasm_file.File))
	return nil
}

func useWasmFunction(wasm_file *WasmFile, key int, value int) error {
	//check if the instance already exists
	if wasmer_module.Instance == nil {
		fmt.Println("Creating Instance...")
		instance, err := wasmcounter.GetNewWasmInstace(env, engine, store, wasm_file.File) //See global variabled
		if err != nil {
			return err
		}
		wasmer_module.Instance = instance
		//TODO: change name from add_one
		smart_contract, err := wasmer_module.Instance.Exports.GetRawFunction("add_one")
		if err != nil {
			return err
		}
		wasmer_module.Function = smart_contract
	}

	fmt.Println(wasmer_module.Function.Type())
	//fmt.Println(smart_contract.ParameterArity())
	//fmt.Println(smart_contract.ResultArity())
	result, err := wasmer_module.Function.Call(key, value)

	if err != nil {
		return err
	}
	//Write to from env to store when some preconditions are met
	mustSaveState()
	nl := convert.SliceOfInt(result)
	key, newVal, oldVal := nl[0], nl[1], nl[2]
	//TODO: Send to ordering service store!
	err = sendToOrdering(key, newVal, oldVal)
	if err != nil {
		fmt.Println(err)
	}
	//TODO: write function that notices about a change to orderingservice
	fmt.Printf("key= %v V=%v N=%v", key, newVal, oldVal)
	fmt.Printf("Returned: %v, Type:%T Store value: %v\n", nl, nl, env.Store)
	return nil
}

func sendToOrdering(key int, newVal int, oldVal int) error {
	body := map[string]int{"Key": key, "NewVal": newVal, "OldVal": oldVal}

	jsonBody, err := json.Marshal(body)
	req, err := http.NewRequest("POST", orderingURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	runtime := &http.Client{}
	res, err := runtime.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
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
	err := LoadState()
	if err != nil {
		panic("Error getting the environment")
	}
	http.HandleFunc("/Add", handlerSet)
	http.HandleFunc("/Upload", handlerUpload)
	//The function embeds ego-certificate on its own
	// tlsConfig, err := enclave.CreateAttestationServerTLSConfig()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	server := http.Server{Addr: ":8086"}
	fmt.Println("Listening...")
	err = server.ListenAndServe()
	//err = server.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Println("Error here!", err)
		return
	}
}

// Stores secrets to disk
//TODO: Below is partly copied code from youtube
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
//If the storage isnt there create one...
func LoadState() error {
	file, err := os.ReadFile("/data/secret.store")
	//if the does not exist...
	if os.IsNotExist(err) {
		//TODO:
		fmt.Println("The file does not exist, creating one in this enclave ...")
		//must save state stores to the store from env
		err = mustSaveState() //In this context it means to create an empty file since Store is empty
		if err != nil {
			return err
		}
		//It is created with sealing key now so we can read it and unseal it
		file, err = os.ReadFile("/data/secret.store")
		if err != nil {
			return err
		}
	}
	//The storage file already exists...
	decrypted_file, err := ecrypto.Unseal(file, nil)
	if err != nil {
		fmt.Println("Error unsealing...")
		return err
	}
	dec := gob.NewDecoder(bytes.NewBuffer(decrypted_file))
	err = dec.Decode(&env.Store)
	fmt.Printf("Store value: %v\n", env.Store)
	if err != nil {
		return err
	}
	return nil
}
