package main

import (
	"bytes"
	"crypto/tls"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/edgelesssys/ego/ecrypto"
	"github.com/gorilla/websocket"
	"github.com/robinaasan/Bachelor_Ego/runtime/handleclient"
	wasmer "github.com/wasmerio/wasmer-go/wasmer"
	//"github.com/edgelesssys/ego/enclave"
)

type Transaction struct {
	handleclient.TransactionContent `json:"TransactionContent"`
}

type BlockFromTransactions struct {
	TransactionContentSlice []handleclient.TransactionContent `json:"TransactionContentSlice"`
	ACK                     string                            `json:"ACK"`
}

// type BlockToTime struct {
// 	CreatedBlock []byte    `json:"CreatedBlock"`
// 	Timer        time.Time `json:"Timer"`
// }

// var timeForResponseSlice []string
// var timeOnSend time.Time

// send the transacions to ordering
func sendToOrdering(setvalues *handleclient.SetValue, nameClient string, tlsConfig *tls.Config, secureURL string, conn *websocket.Conn) error {
	// timeOnSend = time.Now()
	tc := handleclient.TransactionContent{
		ClientName: nameClient,
		Key:        setvalues.Key,
		NewVal:     setvalues.NewVal,
		OldVal:     setvalues.OldVal,
	}
	t := Transaction{
		TransactionContent: tc,
	}
	// fmt.Println(t.TimeStamp)
	jsonBody, err := json.Marshal(t)
	if err != nil {
		return err
	}
	// fmt.Printf("Transaction: %+v\n", t)
	err = conn.WriteMessage(websocket.TextMessage, jsonBody)

	if err != nil {
		log.Println("Write error:", err)
		return err
	}
	return nil
}

// wait for all messages from orderingservice
func WaitForOrderingMessages(conn *websocket.Conn, environment *handleclient.EnvStore) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			// handle error
			log.Println("Write error:", err)
			return
		}
		// if string(message) == "ACK" {
		// 	fmt.Printf("%s\n", message)
		// 	continue
		// }

		blockFromTransactions := &BlockFromTransactions{}
		// BlockToTime := &BlockToTime{}

		err = json.Unmarshal(message, blockFromTransactions)
		// err = json.Unmarshal(message, &BlockToTime)
		// fmt.Printf("%+v", blockFromTransactions.TransactionContentSlice)
		if err != nil {
			panic("Cant read message from orderingservice")
		}
		setTransactionsInEnvironment(blockFromTransactions.TransactionContentSlice, environment)
	}
}

// set the created blocks in the runtime environment
func setTransactionsInEnvironment(transacions []handleclient.TransactionContent, environment *handleclient.EnvStore) error {
	for _, t := range transacions {
		(*environment).Store[int32(t.Key)] = int32(t.NewVal)
		// fmt.Printf("Ready to store: %+v\n", t)
	}
	// store all the transactions
	err := mustSaveState(environment)
	if err != nil {
		return err
	}
	// fmt.Printf("%v\n", environment.Store)
	return nil
}

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

	// // TO ORDERINGSERVICE
	// attestURL := "http://localhost:8087"
	// secureURL := "wss://localhost:443"

	// // create client keys
	// privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	// pubKey := x509.MarshalPKCS1PublicKey(&privKey.PublicKey)

	// // get server certificate over insecure channel
	// serverCert := runtimelocalattestation.HttpGet(nil, attestURL+"/cert")

	// // get the server's report targeted at this client
	// clientInfoReport, err := enclave.GetLocalReport(nil, nil)
	// if err != nil {
	// 	panic(err)
	// }
	// serverReport := runtimelocalattestation.HttpGet(nil, attestURL+"/report", runtimelocalattestation.MakeArg("target", clientInfoReport))

	// // verify server certificate using the server's report
	// if err := verifyreport.VerifyReport(serverReport, serverCert); err != nil {
	// 	panic(err)
	// }

	// // request a client certificate from the server
	// pubKeyHash := sha256.Sum256(pubKey)
	// clientReport, err := enclave.GetLocalReport(pubKeyHash[:], serverReport)
	// if err != nil {
	// 	panic(err)
	// }
	// clientCert := runtimelocalattestation.HttpGet(nil, attestURL+"/client", runtimelocalattestation.MakeArg("pubkey", pubKey), runtimelocalattestation.MakeArg("report", clientReport))

	// // create mutual TLS config
	// tlsConfig := &tls.Config{
	// 	Certificates: []tls.Certificate{
	// 		{
	// 			Certificate: [][]byte{clientCert},
	// 			PrivateKey:  privKey,
	// 		},
	// 	},
	// 	RootCAs: x509.NewCertPool(),
	// }
	// parsedServerCert, _ := x509.ParseCertificate(serverCert)
	// tlsConfig.RootCAs.AddCert(parsedServerCert)

	// // Set the tls config for the runtime
	// tr := &http.Transport{
	// 	TLSClientConfig: tlsConfig,
	// }

	// runtime.SecureRuntimeClient.Transport = tr
	// runtime.SecureRuntimeClient.Timeout = time.Second * 10

	// dialer := websocket.DefaultDialer
	// dialer.TLSClientConfig = tlsConfig

	// conn, _, err := websocket.DefaultDialer.Dial(secureURL+"/transaction", nil)
	// if err != nil {
	// 	panic("something wrong with websocket!")
	// }

	// // set the connection for the runtime
	// runtime.SocketConnectionToOrdering = conn

	// // create a go routine for waiting for messages from the orderingservice
	// go WaitForOrderingMessages(runtime.SocketConnectionToOrdering, runtime.Environment)

	http.HandleFunc("/Init", runtime.InitHandler())
	http.HandleFunc("/Add", runtime.SetHandler(setTransactionsInEnvironment, runtime.Environment))
	http.HandleFunc("/Upload", runtime.UploadHandler())

	// // The function embeds ego-certificate on its own
	// clienttlsConfig, err := enclave.CreateAttestationServerTLSConfig()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// create TLS server for the vendors
	server := http.Server{Addr: ":8086"}
	// err = server.ListenAndServeTLS("", "")
	err = server.ListenAndServe()
	fmt.Println("Listening...")
	if err != nil {
		fmt.Println("Error here!", err)
		return
	}
}

// Stores secrets to disk from the Environment (defined in the wasmcounter package)
func mustSaveState(env *handleclient.EnvStore) error {
	b := new(bytes.Buffer)
	e := gob.NewEncoder(b)
	// Encoding state
	err := e.Encode(env.Store)
	if err != nil {
		return err
	}
	// encrypt the data stored with a key derived from the signer and product id in of the envlave
	encState, err := ecrypto.SealWithProductKey(b.Bytes(), nil)
	if err != nil {
		return err
	}
	if err := os.WriteFile("./secret.store", encState, 0o600); err != nil {
		return fmt.Errorf("Error: creating file responded with: %v", err)
	}
	return nil
}

// read the file and set map in env from storage
// If the storage isn't there create one...
func loadState(env *handleclient.EnvStore) error {
	file, err := os.ReadFile("./secret.store")
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
		file, err = os.ReadFile("./secret.store")
		if err != nil {
			return err
		}
	}
	// unseal the file and read its content
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
