package handleclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/robinaasan/Bachelor_Ego/server/wasmcounter"
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

//Handler for the client
func InitHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		client_name := query.Get("username")
		new_client := NewClient(client_name)
		AllClients[string(new_client.Hash)] = new_client
		fmt.Printf("Createt client with 'hash': %s\n", new_client.Hash)
		fmt.Fprint(w, "ACK\n")
	}
}

func UploadHandler() http.HandlerFunc {
	//TODO: it is the same code as in SetHandler
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		client_name := query.Get("username")
		if client_name == "" {
			fmt.Fprintf(w, "Error: didn't get any username\n")
			return
		}

		theClient, err := GetClient([]byte(client_name))
		if err != nil {
			fmt.Fprintf(w, "Error: couldn't find that client\n")
			return
		}
		err = theClient.GetWasmFile(r)
		if err != nil {
			fmt.Fprint(w, err.Error())
		}
		fmt.Fprint(w, "ACK\n")
	}
}

func SetHandler(mustSaveState func() error, sendToOrdering func(SetValue, string) error, env *wasmcounter.MyEnvironment) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		client_name := query.Get("username")
		if client_name == "" {
			fmt.Fprintf(w, "Error: didn't get any username\n")
		}
		theClient, err := GetClient([]byte(client_name))
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
		//Client use the wasmfunction
		setvalues, err := theClient.UseWasmFunction(key, value, env, wasmcounter.Engine, wasmcounter.Store)
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

		//No error from senToOrdering
		fmt.Fprintf(w, "ACK\n")
	}
}

func Handle_callback(setTransactionsInEnvironment func(*Callback) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		callback := &Callback{}
		err := json.NewDecoder(r.Body).Decode(&callback.CallbackList)
		if err != nil {
			fmt.Fprintf(w, "Error: decoding the data")
			return
		}
		//fmt.Printf("%+v", *callback.CallbackList[0])
		fmt.Fprintf(w, "OK")
		
		setTransactionsInEnvironment(callback)
	}
}

func (cl *Client) GetWasmFile(r *http.Request) error {
	wasmfile := cl.Wasm_file
	err := json.NewDecoder(r.Body).Decode(wasmfile)
	if err != nil {
		return err
	}
	//fmt.Printf("Json: %v", string(cl.Wasm_file.File))
	return nil
}

//Confirm that the client has uploaded a wasm file
func (cl *Client) WasmFileExist() bool {
	return len(cl.Wasm_file.File) != 0
}

func GetClient(hash []byte) (*Client, error) {
	cl, exists := AllClients[string(hash)]
	if exists {
		return cl, nil
	}
	return &Client{}, errors.New("couldnt find any client with that hash.\n")
}
