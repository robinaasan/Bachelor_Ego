package handleclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/wasmerio/wasmer-go/wasmer"
)

type HashResponse struct {
	Hash []byte `json:"Hash"`
}

type Transaction struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	OldVal     int    `json:"OldVal"`
	ClientName string `json:"ClientName"`
}

type Callback struct {
	CallbackList []*Transaction
}

type EnvStore struct {
	Store map[int32]int32
}

type Runtime struct {
	RuntimeClient *http.Client
	Engine        *wasmer.Engine
	WasmStore     *wasmer.Store
	Environment   *EnvStore
	AllClients    AllClients
}

// Handler for the client
func (runtime *Runtime) InitHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		client_name := query.Get("username")
		new_client := NewClient(client_name)
		runtime.AllClients[string(new_client.Hash)] = new_client

		fmt.Printf("Createt client with 'hash': %s\n", new_client.Hash)
		fmt.Fprint(w, "ACK")
	}
}

func (runtime *Runtime) UploadHandler() http.HandlerFunc {
	// TODO: it is the same code as in SetHandler
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		client_name := query.Get("username")
		if client_name == "" {
			fmt.Fprintf(w, "Error: didn't get any username")
			return
		}

		theClient, err := GetClient([]byte(client_name), runtime.AllClients)
		if err != nil {
			fmt.Fprintf(w, "Error: couldn't find that client")
			return
		}
		err = theClient.GetWasmFile(r)
		if err != nil {
			fmt.Fprint(w, err.Error())
		}
		fmt.Fprint(w, "ACK")
	}
}

func (runtime *Runtime) SetHandler(sendToOrdering func(SetValue, string) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		client_name := query.Get("username")
		if client_name == "" {
			fmt.Fprintf(w, "Error: didn't get any username\n")
		}
		theClient, err := GetClient([]byte(client_name), runtime.AllClients)
		if err != nil {
			fmt.Fprintf(w, "Error: getting the client\n")
			return
		}

		if !theClient.WasmFileExist() {
			fmt.Fprintf(w, "Error: now wasm module uploaded")
			return
		}

		var key, value int
		key, err = strconv.Atoi(query.Get("key"))
		if err != nil {
			fmt.Fprintf(w, "Error: couldn't get the key\n")
			return
		} else {
			value, err = strconv.Atoi(query.Get("value"))
			if err != nil {
				fmt.Fprintf(w, "Error: couldn't get the value\n")
				return
			}
		}
		// Client use the wasmfunction
		setvalues, err := theClient.UseWasmFunction(key, value, runtime)
		if err != nil {
			fmt.Println(err)
			fmt.Fprintln(w, err)
			return
		}
		err = sendToOrdering(setvalues, string(theClient.Hash))
		if err != nil {
			fmt.Printf("Error sending to orderingservice: %s", err.Error())
			return
		}

		// No error from sendToOrdering
		fmt.Fprintf(w, "ACK")
	}
}

func (runtime *Runtime) Handle_callback(mustSaveState func(*EnvStore) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		callback := &Callback{}
		err := json.NewDecoder(r.Body).Decode(&callback.CallbackList)
		if err != nil {
			fmt.Fprintf(w, "Error: decoding the data")
			return
		}
		// fmt.Printf("%+v", *callback.CallbackList[0])
		// fmt.Fprintf(w, "OK")

		runtime.setTransactionsInEnvironment(mustSaveState, callback)
	}
}

func (runtime *Runtime) setTransactionsInEnvironment(mustSaveState func(*EnvStore) error, c *Callback) error {
	for _, t := range c.CallbackList {
		(*runtime.Environment).Store[int32(t.Key)] = int32(t.NewVal)
	}
	// store all the transactions
	err := mustSaveState(runtime.Environment)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", runtime.Environment.Store)
	return nil
}

func (cl *Client) GetWasmFile(r *http.Request) error {
	wasmfile := cl.Wasm_file
	err := json.NewDecoder(r.Body).Decode(wasmfile)
	if err != nil {
		return err
	}
	// fmt.Printf("Json: %v", string(cl.Wasm_file.File))
	return nil
}

// Confirm that the client has uploaded a wasm file
func (cl *Client) WasmFileExist() bool {
	return len(cl.Wasm_file.File) != 0
}

func GetClient(hash []byte, allClients AllClients) (*Client, error) {
	cl, exists := allClients[string(hash)]
	if exists {
		return cl, nil
	}
	return &Client{}, errors.New("couldnt find any client with that hash.\n")
}
