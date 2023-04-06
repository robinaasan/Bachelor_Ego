package main

import (
	"crypto"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edgelesssys/ego/enclave"
	"github.com/gorilla/websocket"
	"github.com/robinaasan/Bachelor_Ego/orderingservice/blockchain"
	"github.com/robinaasan/Bachelor_Ego/orderingservice/orderinglocalattestation"
	"github.com/robinaasan/Bachelor_Ego/orderingservice/runtimeclients"
	"github.com/robinaasan/Bachelor_Ego/verifyreport"
)

const (
	PATH      = "./blockFiles/"
	genesis   = "Block1.json"
	blockSize = 5
)

type ResponsesRuntime struct {
	response string
	endpoint string
	err      error
}

type BlockTransactionStore struct {
	allTransactions      []runtimeclients.Transaction
	blockchain           *blockchain.BlockChain
	count                int
	client               *http.Client
	runtime_clients      []runtimeclients.Runtimeclient
	dataToSendInCallback []byte
	mu                   *sync.Mutex
}

// type handleConcurrentRequests struct {
// 	mu          sync.Mutex
// 	wg          sync.WaitGroup
// 	cond        *sync.Cond
// 	runCallback bool
// }

func main() {
	// TODO: verify the integrity of the blocks if there is a genesis block
	genBlock := fmt.Sprintf("%s%s", PATH, genesis)
	block_chain := blockchain.InitBlockChain(time.Now().String())
	//blockSice := 5
	//create the condition in handleConcurrentRequest
	//handleConcurrentRequests := handleConcurrentRequests{}
	//handleConcurrentRequests.cond = sync.NewCond(&handleConcurrentRequests.mu)

	//create the blocktransactionstore
	blockTransactionStore := BlockTransactionStore{blockchain: block_chain, count: 0}
	if !fileExist(genBlock) {
		err := addBlockFile(genBlock, blockTransactionStore.blockchain.Blocks[0])
		if err != nil {
			fmt.Println(err)
		}
	} else { // Load the rest of the blockchain
		err := ReadAllBlockFiles(&blockTransactionStore)
		if err != nil {
			fmt.Println(err)
		}
	}
	// create the server certificate and the servers
	cert, privKey := orderinglocalattestation.CreateServerCertificate()
	attestServer := newAttestServer(cert, privKey)
	//create upgrader for websocket
	var upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	blockTransactionStore.runtime_clients = []runtimeclients.Runtimeclient{}

	createdBlock := make(chan []byte)
	go blockTransactionStore.waitForBlockFromTransactions(createdBlock)
	secureServer := newSecureServer(cert, privKey, &blockTransactionStore, upgrader, createdBlock)

	// run the servers
	go func() {
		err := attestServer.ListenAndServe()
		panic(err)
	}()

	fmt.Println("listening ...")
	err := secureServer.ListenAndServeTLS("", "")

	if err != nil {
		fmt.Println(err)
	}

	//blockTransactionStore.blockchain.PrintChain()
	//http.HandleFunc("/", blockTransactionStore.handlerTransaction(blockSize))
	//server := http.Server{Addr: "localhost:8087"}
	//err := server.ListenAndServe()
	//fmt.Println(err)
}

func newAttestServer(cert []byte, privKey crypto.PrivateKey) *http.Server {
	certHash := sha256.Sum256(cert)
	mux := http.NewServeMux()

	// Returns the server certificate.
	mux.HandleFunc("/cert", func(w http.ResponseWriter, r *http.Request) { w.Write(cert) })

	// Returns a local report including the server certificate's hash for the given target report.
	mux.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		targetReport := orderinglocalattestation.GetQueryArg(w, r, "target")
		if targetReport == nil {
			return
		}
		report, err := enclave.GetLocalReport(certHash[:], targetReport)
		if err != nil {
			http.Error(w, fmt.Sprintf("GetLocalReport: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write(report)
	})

	// Returns a client certificate for the given pubkey.
	// The given report ensures that only verified enclaves can get certificates for their pubkeys.
	mux.HandleFunc("/client", func(w http.ResponseWriter, r *http.Request) {
		pubKey := orderinglocalattestation.GetQueryArg(w, r, "pubkey")
		if pubKey == nil {
			return
		}
		report := orderinglocalattestation.GetQueryArg(w, r, "report")
		if report == nil {
			return
		}
		if err := verifyreport.VerifyReport(report, pubKey); err != nil {
			http.Error(w, fmt.Sprintf("verifyReport: %v", err), http.StatusBadRequest)
			return
		}
		w.Write(orderinglocalattestation.CreateClientCertificate(pubKey, cert, privKey))
	})

	return &http.Server{
		Addr:    "localhost:8087",
		Handler: mux,
	}
}

func newSecureServer(cert []byte, privKey crypto.PrivateKey, bt *BlockTransactionStore, upgrader *websocket.Upgrader, createdBlock chan []byte) *http.Server {
	mux := http.NewServeMux()
	//mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "pong") })
	mux.HandleFunc("/transaction", bt.handlerTransaction(blockSize, upgrader, createdBlock))
	//mux.HandleFunc("/callback", bt.handlerCallback())
	// use server certificate also as client CA
	parsedCert, _ := x509.ParseCertificate(cert)
	clientCAs := x509.NewCertPool()
	clientCAs.AddCert(parsedCert)

	return &http.Server{
		Addr:    "localhost:443",
		Handler: mux,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{
				{
					Certificate: [][]byte{cert},
					PrivateKey:  privKey,
				},
			},
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  clientCAs,
		},
	}
}

// Add the block to the blockChain
// TODO: notify the runtimes about the change!
func (bt *BlockTransactionStore) handlerTransaction(blockSice int, upgrader *websocket.Upgrader, createdBlock chan []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//start := time.Now()

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading websocket:", err)
			return
		}
		//defer conn.Close()
		newClient := &runtimeclients.Runtimeclient{
			Conn: conn,
			Send: make(chan []byte),
		}
		bt.runtime_clients = append(bt.runtime_clients, *newClient)

		// start a goroutine to handle receiving messages from this client

		//go newClient.ReadPump(bt.count, mu, &bt.allTransactions, createdBlock)

		go newClient.ReadPump(bt.count, &bt.allTransactions, createdBlock)

		go newClient.WritePump()
		// block_chain.AddNewblock(transactionData, time.Now().String(), clientName)
		fmt.Println("Ready for callback")
		// for {
		// 	_, msg, err := conn.ReadMessage()

		// 	bt.handleConcurrentRequests.mu.Unlock()

		// }
		//fmt.Printf("%v ms elapsed\n", time.Since(start).Microseconds())
		// fmt.Printf("%.4fms elapsed", time.Since(start).Milliseconds())
		//fmt.Fprintf(w, "ACK")
		// s := fmt.Sprintf("%s", r.RemoteAddr)
	}
}

func (bt *BlockTransactionStore) waitForBlockFromTransactions(createdBlockbytes chan []byte) {
	//c := <- createdBlockbytes
	for {
		select {
		case c := <-createdBlockbytes:
			bt.blockchain.AddNewblock(c, time.Now().String())
			addedBlock := bt.blockchain.Blocks[len(bt.blockchain.Blocks)-1]
			newBlockFileName := fmt.Sprintf("%s%s.json", PATH, fmt.Sprintf("Block%v", len(bt.blockchain.Blocks)))
			err := addBlockFile(newBlockFileName, addedBlock)
			if err != nil {
				//TODO: handle error
				return
			}
			fmt.Println("Should broadcast...")
			runtimeclients.BroadcastMessage(c, bt.runtime_clients)
		}
	}
}

// func (bt *BlockTransactionStore) handlerCallback() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		//wait for the block to be created
// 		bt.handleConcurrentRequests.cond.L.Lock()
// 		//for !bt.handleConcurrentRequests.runCallback {
// 		fmt.Println("waiting for the flag to be set to true...")
// 		bt.handleConcurrentRequests.cond.Wait()
// 		//}
// 		//bt.handleConcurrentRequests.mu.Lock()
// 		fmt.Println("Sending back block...")
// 		_, err := w.Write(bt.dataToSendInCallback)
// 		if err != nil {
// 			panic("Error sending the data to runtimes")
// 		}
// 		//bt.handleConcurrentRequests.mu.Unlock()
// 		//bt.handleConcurrentRequests.runCallback = false
// 		bt.handleConcurrentRequests.cond.L.Unlock()

// 		//respond with the block generated

// 	}
// }

// Add the block as a json file in the filesystem
func addBlockFile(filename string, b *blockchain.Block) error {
	jsonBody, err := b.Serialize()
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, jsonBody, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func fileExist(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func ReadAllBlockFiles(blockTransactionStore *BlockTransactionStore) error {
	files, err := os.ReadDir(PATH)
	if err != nil {
		return err
	}
	for _, file_entry := range files {
		fileType := strings.Split(file_entry.Name(), ".")
		if fileType[1] != "json" {
			return errors.New("wrong file type\n.")
		}
		fmt.Println(file_entry.Name())
		newBlock := &blockchain.Block{}
		fileBytes, err := os.ReadFile(PATH + file_entry.Name())
		if err != nil {
			return err
		}
		err = json.Unmarshal(fileBytes, newBlock)
		if err != nil {
			return err
		}
		// TODO: now the genesys block changes gets the the date updated i the blockchain, it is not created a new one
		// There is probably a better solution than this
		// if it is the genesis file create that first
		if fileType[0] == "Block1" {
			// The genesis block was created in main
			// Below we use the timestamp and set the same hash as is stored
			(*blockTransactionStore).blockchain.Blocks[0] = blockchain.CreateGenesis(newBlock.TimeStamp)
			(*blockTransactionStore).blockchain.Blocks[0].Data = newBlock.Data
		} else { // genesis block is already created in the filesystem
			(*blockTransactionStore).blockchain.AddNewblock(newBlock.Data, newBlock.TimeStamp)
		}

	}
	return nil
}

// func (bt *BlockTransactionStore) sendCallback(allTransactionBytes []byte, endpoints []string) {
// 	// var wg sync.WaitGroup
// 	c := make(chan ResponsesRuntime)
// 	for _, endpoint := range endpoints {
// 		bt.handleConcurrentRequests.wg.Add(1)
// 		go checkURL(endpoint, c, &bt.handleConcurrentRequests.wg, allTransactionBytes, bt.client)
// 	}
// 	go func() {
// 		bt.handleConcurrentRequests.wg.Wait()
// 		close(c)
// 	}()

// 	// for r := range c {
// 	// 	// if r.err != nil {

// 	// 	// 	s := fmt.Sprintf("Error: endpoint: %s got: %v\n", r.endpoint, r.err)
// 	// 	// 	fmt.Printf("%v", s)
// 	// 	// } else {
// 	// 	// 	fmt.Println(r.response + "\n")
// 	// 	// }

// 	// 	// if r.err != nil {
// 	// 	// 	fmt.Printf("Error requesting %s: %v\n", r.endpoint, r.err)
// 	// 	// 	continue
// 	// 	// }
// 	// 	fmt.Printf("%+v\n", r)
// 	// }
// }

// func checkURL(endpoint string, c chan ResponsesRuntime, wg *sync.WaitGroup, allTransactionBytes []byte, cl *http.Client) {
// 	defer (*wg).Done()

// 	// responseruntime := ResponsesRuntime{endpoint: endpoint}
// 	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(allTransactionBytes))
// 	// if err != nil {
// 	// 	s = err.Error()
// 	// }
// 	if err != nil {
// 		c <- ResponsesRuntime{endpoint, "", err}
// 		return
// 	}
// 	req.Header.Set("Content-Type", "application/json")

// 	res, err := cl.Do(req)
// 	if err != nil {
// 		c <- ResponsesRuntime{endpoint, "", err}
// 		return
// 	}
// 	defer res.Body.Close()
// 	// resBody, err := io.ReadAll(res.Body)

// 	// fmt.Printf("Res: %v", string(resBody))
// 	c <- ResponsesRuntime{endpoint, res.Status, nil}
// }
