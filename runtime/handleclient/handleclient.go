package handleclient

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/wasmerio/wasmer-go/wasmer"
)

type EnvStore struct {
	Store map[int32]int32
}

type AllClients map[string]*Client

type Runtime struct {
	sync.Mutex
	SecureRuntimeClient        *http.Client
	Engine                     *wasmer.Engine
	WasmStore                  *wasmer.Store
	Environment                *EnvStore
	AllClients                 AllClients
	TlsConfig                  *tls.Config
	SocketConnectionToOrdering *websocket.Conn
	Timeout                    time.Duration
}

// Handler for the client/vendor
func (runtime *Runtime) InitHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		client_name := query.Get("username")
		_, err := GetClient(client_name, runtime.AllClients)
		if err == nil {
			fmt.Fprint(w, "This client already exists")
			return
		}

		new_client := NewClient(client_name)
		runtime.AllClients[string(new_client.Hash)] = new_client
		new_client.ClientMessages = make(map[string]bool)
		new_client.WaitForAckFromOrdering = make(chan string)

		fmt.Printf("Createt client with 'hash': %s\n", new_client.Hash)
		fmt.Fprint(w, "ACK")

	}
}

func (runtime *Runtime) UploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		client_name := query.Get("username")
		if client_name == "" {
			fmt.Fprintf(w, "Error: didn't get any username")
			return
		}

		theClient, err := GetClient(client_name, runtime.AllClients)
		if err != nil {
			fmt.Fprintf(w, "Error: couldn't find the client")
			return
		}

		if theClient.WasmFileExist() {
			fmt.Fprint(w, "Uploading a new wasm module...")
		}

		// set the wasm module
		err = theClient.SetWasmFile(r)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}

		// create the instance for the vendor
		err = theClient.CreateInstanceClient(runtime)
		if err != nil {
			fmt.Fprint(w, err.Error())
		} else {
			fmt.Fprint(w, "ACK")
		}
	}
}

func (runtime *Runtime) SetHandler(sendToOrdering func(*SetValue, *Client, string, *tls.Config, string, *websocket.Conn) error, secureURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		timer := time.NewTimer(runtime.Timeout)

		runtime.Lock()
		defer runtime.Unlock()
		query := r.URL.Query()

		// check that the user exists
		client_name := query.Get("username")
		if client_name == "" {
			fmt.Fprintf(w, "Error: didn't get any username\n")
			return
		}

		theClient, err := GetClient(client_name, runtime.AllClients)
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
		messageId := uuid.New().String()
		theClient.ClientMessages[messageId] = true
		//add to client in a chan to indicate that there is a message waiting for ack
		//runtime.ClientMessageChan <- theClient
		err = sendToOrdering(setvalues, theClient, messageId, runtime.TlsConfig, secureURL, runtime.SocketConnectionToOrdering)
		if err != nil {
			fmt.Printf("Error sending to orderingservice: %s", err.Error())
			return
		}
		select {
		case clientmessage := <-theClient.WaitForAckFromOrdering:
			fmt.Fprintln(w, clientmessage)
		case <-timer.C:
			fmt.Fprintln(w, "")
		}
	}
}
