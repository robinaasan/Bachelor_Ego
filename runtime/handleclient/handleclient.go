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
		clientName := query.Get("username")
		cl := GetClient(clientName, runtime.AllClients)
		if cl.Hash != nil {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "that client already exists")
			return
		}
		// create the new client
		newClient := NewClient(clientName)
		// add the client to the AllClients map
		runtime.AllClients[string(newClient.Hash)] = newClient

		fmt.Printf("Createt client with 'hash': %s\n", newClient.Hash)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ACK")

	}
}

func (runtime *Runtime) UploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientHasWasmMod := false
		query := r.URL.Query()
		clientName := query.Get("username")
		if clientName == "" {
			http.Error(w, "didn't get any username", http.StatusForbidden)
			return
		}

		theClient := GetClient(clientName, runtime.AllClients)
		if theClient.Hash == nil {
			http.Error(w, "couldn't find the client", http.StatusForbidden)
			return
		}

		if theClient.WasmFileExist() {
			clientHasWasmMod = true
		}
		// set the wasm module
		err := theClient.SetWasmFile(r)
		if err != nil {
			fmt.Println("Error: ", err.Error())
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		// create the instance for the vendor
		err = theClient.CreateInstanceClient(runtime)
		if err != nil {
			fmt.Fprint(w, err.Error())
		} else {
			w.WriteHeader(http.StatusOK)
			if clientHasWasmMod {
				fmt.Fprint(w, "Reuploaded new wasm module")
			} else {
				fmt.Fprint(w, "ACK")
			}
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
		clientName := query.Get("username")
		if clientName == "" {
			http.Error(w, "Issue with the url values", http.StatusForbidden)
			return
		}

		theClient := GetClient(clientName, runtime.AllClients)
		if theClient.Hash == nil {
			http.Error(w, "Could not find that client", http.StatusForbidden)
			return
		}

		if !theClient.WasmFileExist() {
			http.Error(w, "No Wasm found for that client", http.StatusForbidden)
			return
		}

		var key, value int
		key, err := strconv.Atoi(query.Get("key"))
		if err != nil {
			fmt.Fprintf(w, "Error: couldn't get the key\n")
			return
		} else {
			value, err = strconv.Atoi(query.Get("value"))
			if err != nil {
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}
		}
		// Client use the wasmfunction
		setvalues, err := theClient.UseWasmFunction(key, value, runtime)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		messageId := uuid.New().String()
		theClient.ClientMessages[messageId] = true
		//add to client in a chan to indicate that there is a message waiting for ack
		//runtime.ClientMessageChan <- theClient
		err = sendToOrdering(setvalues, theClient, messageId, runtime.TlsConfig, secureURL, runtime.SocketConnectionToOrdering)
		if err != nil {
			fmt.Printf("Error sending to orderingservice: %s\n", err.Error())
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		// wait for ack from orderingservice, or timout ...
		select {
		case clientMessage := <-theClient.WaitForAckFromOrdering:
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, clientMessage)
		case <-timer.C:
			http.Error(w, "Request times out", http.StatusRequestTimeout)
		}
	}
}
