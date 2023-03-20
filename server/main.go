package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/edgelesssys/ego/ecrypto"
	"github.com/robinaasan/Bachelor_Ego/server/handleclient"
	wasmer "github.com/wasmerio/wasmer-go/wasmer"
	//"github.com/edgelesssys/ego/enclave"
)

const orderingURL = "http://localhost:8087"

func sendToOrdering(setvalues handleclient.SetValue, nameClient string) error {
	t := handleclient.Transaction{
		ClientName: nameClient,
		Key:        setvalues.Key,
		NewVal:     setvalues.NewVal,
		OldVal:     setvalues.OldVal,
	}
	// q := url.Values{}
	// body := map[string]int{"Key": setvalues.Key, "NewVal": setvalues.NewVal, "OldVal": setvalues.OldVal}
	// q.Add("client", nameClient)
	jsonBody, err := json.Marshal(t)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", orderingURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	// req.URL.RawQuery = q.Encode()
	runtime := &http.Client{}
	res, err := runtime.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(responseData))
	return nil
}

// TODO:
// When a client uploads a smart contract he will be verified by TLS
// Find a way to differanciate the clients and not just use the name...

//TODO: create new endpoint for adding the new client
//Return the address for where this client is stored to the client (Connection with TLS) so it can be used in the UPLOAD and SET handlers
//-Create a new package for the clients where name and wasm_files are stored
//-Create the handlers in that package in own file
//-In this main file have the code for contacting the ordering service
func main() {
	eng := wasmer.NewEngine()
	runtime := &handleclient.Runtime{
		RuntimeClient: &http.Client{},
		Engine:        eng,
		WasmStore:     wasmer.NewStore(eng),
		Environment:   &handleclient.EnvStore{Store: make(map[int32]int32)},
		AllClients:    make(map[string]*handleclient.Client),
	}
	err := loadState(runtime.Environment)
	if err != nil {
		panic("Error getting the environment")
	}

	http.HandleFunc("/Init", runtime.InitHandler())
	http.HandleFunc("/Add", runtime.SetHandler(sendToOrdering))
	http.HandleFunc("/Upload", runtime.UploadHandler())
	http.HandleFunc("/Callback", runtime.Handle_callback(mustSaveState))
	// TODO: get response from senToOrdering and call handle_callback()
	// The function embeds ego-certificate on its own
	// tlsConfig, err := enclave.CreateAttestationServerTLSConfig()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	server := http.Server{Addr: ":8086"}
	fmt.Println("Listening...")
	err = server.ListenAndServe()
	// err = server.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Println("Error here!", err)
		return
	}
}

// Stores secrets to disk from the Environment (defined in the wasmcounter package)
// TODO: Below is partly copied code from youtube
func mustSaveState(env *handleclient.EnvStore) error {
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
	if err := os.WriteFile("/data/secret.store", encState, 0o600); err != nil {
		return fmt.Errorf("Error: creating file responded with: %v", err)
	}
	return nil
}

// read the file and set map in env from storage
// If the storage isnt there create one...
func loadState(env *handleclient.EnvStore) error {
	file, err := os.ReadFile("/data/secret.store")
	// if the does not exist...
	if os.IsNotExist(err) {
		// TODO:
		fmt.Println("The file does not exist, creating one in this enclave ...")
		// must save state stores to the store from env
		err = mustSaveState(env) // In this context it means to create an empty file since Store is empty
		if err != nil {
			return err
		}
		// It is created with sealing key now so we can read it and unseal it
		file, err = os.ReadFile("/data/secret.store")
		if err != nil {
			return err
		}
	}
	// The storage file already exists...
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
