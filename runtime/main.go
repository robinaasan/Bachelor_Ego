package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/edgelesssys/ego/ecrypto"
	"github.com/edgelesssys/ego/enclave"
	"github.com/gorilla/websocket"
	"github.com/robinaasan/Bachelor_Ego/runtime/handleclient"
	"github.com/robinaasan/Bachelor_Ego/runtime/runtimelocalattestation"
	"github.com/robinaasan/Bachelor_Ego/verifyreport"
	wasmer "github.com/wasmerio/wasmer-go/wasmer"
	//"github.com/edgelesssys/ego/enclave"
)

type Transaction struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	OldVal     int    `json:"OldVal"`
	ClientName string `json:"ClientName"`
}

type Callback struct {
	CallbackList []*Transaction
}

func sendToOrdering(setvalues handleclient.SetValue, nameClient string, tlsConfig *tls.Config, secureURL string, conn *websocket.Conn) error {
	t := Transaction{
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
	// req, err := http.NewRequest("POST", orderingURL, bytes.NewBuffer(jsonBody))
	// if err != nil {
	// 	return err
	// }
	// req.Header.Add("Content-Type", "application/json")

	// req.URL.RawQuery = q.Encode()
	// use the established secure channel

	err = conn.WriteMessage(websocket.TextMessage, jsonBody)

	if err != nil {
		log.Println("Write error:", err)
		return err
	}
	// Read message from server
	// _, message, err := conn.ReadMessage()
	// if err != nil {
	// 	log.Println("Read error:", err)
	// 	return err
	// }

	// // Print received message
	// fmt.Printf("Received message: %s\n", message)

	// resp, err := runtimelocalattestation.HttpPost(tlsConfig, jsonBody, secureURL+"/transaction")
	// if err != nil {
	// 	fmt.Printf("Error sending to ordering: %v", err)
	// 	return err
	// }
	// fmt.Printf("server responded: %s\n", resp)
	// //res, err := runtime.Do(req)
	// // if err != nil {
	// // 	return err
	// // }
	// // defer res.Body.Close()
	// // responseData, err := io.ReadAll(res.Body)
	// // if err != nil {
	// // 	return err
	// // }
	// fmt.Println(string(resp))
	return nil
}

func WaitForOrderingMessages(conn *websocket.Conn, environment *handleclient.EnvStore) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			// handle error
			return
		}
		if string(message) == "ACK" {
			fmt.Printf("%s\n", message)
			continue
		}
		callback := &Callback{}
		err = json.Unmarshal(message, &callback.CallbackList)
		if err != nil {
			panic("Cant read message from orderingservice")
		}
		setTransactionsInEnvironment(callback, environment)
		//if messageType == websocket.TextMessage

	}
}

func setTransactionsInEnvironment(c *Callback, environment *handleclient.EnvStore) error {
	for _, t := range c.CallbackList {
		(*environment).Store[int32(t.Key)] = int32(t.NewVal)
		fmt.Printf("Received message: %+v\n", t)
	}
	// store all the transactions
	err := mustSaveState(environment)
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", environment.Store)
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
		SecureRuntimeClient: &http.Client{},
		Engine:              eng,
		WasmStore:           wasmer.NewStore(eng),
		Environment:         &handleclient.EnvStore{Store: make(map[int32]int32)},
		AllClients:          make(map[string]*handleclient.Client),
	}
	err := loadState(runtime.Environment)
	if err != nil {
		panic("Error getting the environment")
	}

	//TO ORDERINGSERVICE
	var attestURL = "http://localhost:8087"
	var secureURL = "wss://localhost:443"

	// create client keys
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	pubKey := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)

	// get server certificate over insecure channel
	serverCert := runtimelocalattestation.HttpGet(nil, attestURL+"/cert")

	// get the server's report targeted at this client
	clientInfoReport, err := enclave.GetLocalReport(nil, nil)
	if err != nil {
		panic(err)
	}
	serverReport := runtimelocalattestation.HttpGet(nil, attestURL+"/report", runtimelocalattestation.MakeArg("target", clientInfoReport))

	// verify server certificate using the server's report
	if err := verifyreport.VerifyReport(serverReport, serverCert); err != nil {
		panic(err)
	}

	// request a client certificate from the server
	pubKeyHash := sha256.Sum256(pubKey)
	clientReport, err := enclave.GetLocalReport(pubKeyHash[:], serverReport)
	if err != nil {
		panic(err)
	}
	clientCert := runtimelocalattestation.HttpGet(nil, attestURL+"/client", runtimelocalattestation.MakeArg("pubkey", pubKey), runtimelocalattestation.MakeArg("report", clientReport))

	// create mutual TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{clientCert},
				PrivateKey:  privKey,
			},
		},
		RootCAs: x509.NewCertPool(),
	}
	parsedServerCert, _ := x509.ParseCertificate(serverCert)
	tlsConfig.RootCAs.AddCert(parsedServerCert)
	//Set the tls config for the runtime

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	runtime.SecureRuntimeClient.Transport = tr
	runtime.SecureRuntimeClient.Timeout = time.Second * 10

	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = tlsConfig

	conn, _, err := websocket.DefaultDialer.Dial(secureURL+"/transaction", nil)
	if err != nil {
		panic("something wrong with websocket!")
	}
	runtime.SocketConnectionToOrdering = conn

	//TODO:
	//SECURE CHANNAL SHOULD BE ESTABLISHED
	//WAIT FOR ALL TRANSACTIONS:
	// go func() {
	// 	err := runtime.Handle_callback(mustSaveState, secureURL+"/callback")
	// 	panic(err)
	// }()
	//ENDORDERINGSERVER
	http.HandleFunc("/Init", runtime.InitHandler(WaitForOrderingMessages))
	http.HandleFunc("/Add", runtime.SetHandler(sendToOrdering, secureURL))
	http.HandleFunc("/Upload", runtime.UploadHandler())
	//http.HandleFunc("/Callback", runtime.Handle_callback(mustSaveState))

	// TODO: get response from senToOrdering and call handle_callback()
	// The function embeds ego-certificate on its own
	clienttlsConfig, err := enclave.CreateAttestationServerTLSConfig()
	if err != nil {
		log.Fatal(err)
	}
	server := http.Server{Addr: ":8086", TLSConfig: clienttlsConfig}
	//err = server.ListenAndServe()
	err = server.ListenAndServeTLS("", "")
	fmt.Println("Listening...")
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
