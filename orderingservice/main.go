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
	PATH      = "./files/blockFiles/"
	genesis   = "000Block1.json"
	blockSize = 5
)

type BlockTransactionStore struct {
	allTransactions []runtimeclients.TransactionContent //Slice of all transactions to be cerated as a block
	blockchain      *blockchain.BlockChain              //Blockchain from the blockchain module
	runtime_clients []runtimeclients.Runtimeclient      //Slice of all connected runtimes
	mu              sync.Mutex
}

func main() {
	genBlock := fmt.Sprintf("%s%s", PATH, genesis) // path to the genesis block in the filesystem
	// Initialize the blockchain
	block_chain := blockchain.InitBlockChain(time.Now().String())
	// create the blocktransactionstore with the created blockchain
	blockTransactionStore := &BlockTransactionStore{blockchain: block_chain, runtime_clients: []runtimeclients.Runtimeclient{}}

	// check if the genesis block is already created if not create it
	// Assuming there are not any blocks if the genesis block does not exists
	if !fileExist(genBlock) {
		err := addBlockFile(genBlock, &blockTransactionStore.blockchain.Blocks[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		// Genesis block exists, load the rest of the blockchain
	} else {
		err := readAllBlockFiles(blockTransactionStore.blockchain)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	// create the server certificate and the servers
	cert, privKey := orderinglocalattestation.CreateServerCertificate()
	attestServer := newAttestServer(cert, privKey)

	// create upgrader for websocket
	var upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// channel for sending the created blocks to the runtimes, sending to this channel depends on the blockSize constant
	blockFromTransactions := make(chan runtimeclients.BlockFromTransactions)
	// go routine for waiting for the blocks to be created
	go blockTransactionStore.waitForBlockFromRuntimeTransactions(blockFromTransactions)

	// create the secure server in the orderingservice
	secureServer := newSecureServer(cert, privKey, blockTransactionStore, upgrader, blockFromTransactions)
	// run the servers
	go func() {
		err := attestServer.ListenAndServe()
		panic(err)
	}()

	fmt.Println("listening ...")
	err := secureServer.ListenAndServeTLS("", "")

	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
}

// unsecure server
func newAttestServer(cert []byte, privKey crypto.PrivateKey) *http.Server {
	// create hash from the serverCertificate.
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
	// The given report ensures that only verified enclaves (runtimes) can get certificates for their pubkeys.
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

// create the secure server
func newSecureServer(cert []byte, privKey crypto.PrivateKey, bt *BlockTransactionStore, upgrader *websocket.Upgrader, blockFromTransactions chan runtimeclients.BlockFromTransactions) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/transaction", bt.handlerTransaction(blockSize, upgrader, blockFromTransactions))

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

// endpoint for handling the transactions from the verified runtimes
func (bt *BlockTransactionStore) handlerTransaction(blockSize int, upgrader *websocket.Upgrader, blockFromTransactions chan runtimeclients.BlockFromTransactions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error upgrading websocket:", err)
			panic(err)
		}
		// create the connected runtimeccclient
		newClient := &runtimeclients.Runtimeclient{
			Conn: conn,
			Send: make(chan runtimeclients.SendBackToRuntime, 1),
		}

		bt.runtime_clients = append(bt.runtime_clients, *newClient)
		// start a goroutine to handle receiving messages from this client(runtime)
		go newClient.ReadPump(blockSize, &bt.allTransactions, &bt.mu, blockFromTransactions)
		// go routine for writing messages to the client(runtime)
		go newClient.WritePump()
		fmt.Println("Succesfully connected to new runtime...")
	}
}

func (bt *BlockTransactionStore) waitForBlockFromRuntimeTransactions(blockFromTransactions chan runtimeclients.BlockFromTransactions) {
	for {
		select {
		case c := <-blockFromTransactions:
			// check if the message is to be broadcasted to all runtimes
			// add the block from all the transactions
			// send block to all the runtimes
			//else
			// send ack to runtime (empty blockslice)

			if c.BroadcastToRuntimes {
				allTransactionsData, err := json.Marshal(c.TransactionContentSlice)
				if err != nil {
					panic("Couldnt marshal the transactions")
				}
				bt.blockchain.AddNewblock(allTransactionsData, time.Now().String())
				addedBlock := &bt.blockchain.Blocks[len(bt.blockchain.Blocks)-1]
				// filename is UnixNano so it gets stored sequentially
				newBlockFileName := fmt.Sprintf("%s%v.json", PATH, time.Now().UnixNano())
				err = addBlockFile(newBlockFileName, addedBlock)
				if err != nil {
					panic("cant store file(s) in the file system")
				}

				//send the block to all connected runtimes
				//fmt.Println("Client was:", c.ClientHash)
				runtimeclients.BroadcastMessage(&runtimeclients.SendBackToRuntime{TransactionContentSlice: c.TransactionContentSlice, ACK: false, MessageId: c.MessageId, ClientHash: c.ClientHash}, bt.runtime_clients)
			} else { //send only an ack to the runtime
				runtimeclients.BroadcastMessage(&runtimeclients.SendBackToRuntime{TransactionContentSlice: []runtimeclients.TransactionContent{}, ACK: true, MessageId: c.MessageId, ClientHash: c.ClientHash}, []runtimeclients.Runtimeclient{*c.Runtimeclient})
			}
		}
	}
}

// Add the block as a json file in the filesystem
func addBlockFile(filename string, b *blockchain.Block) error {
	jsonBody, err := b.Serialize()
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, jsonBody, 0o777)
	if err != nil {
		return err
	}
	return nil
}

// check if a specific file (block) exists in the filesystem
func fileExist(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// load all blocks from the filesystem
func readAllBlockFiles(block_chain *blockchain.BlockChain) error {
	files, err := os.ReadDir(PATH)
	if err != nil {
		return err
	}
	for _, file_entry := range files {
		fileType := strings.Split(file_entry.Name(), ".")
		if fileType[1] != "json" {
			return errors.New("found wrong file type for file")
		}
		fmt.Println("Loading file ", file_entry.Name(), "...")
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
		if fileType[0] == "000Block1" {
			// The genesis block was created in main
			// Below we use the timestamp from the newBlock and create a "new" genesisblock which will have the same hash due to Sha256
			(*block_chain).Blocks[0] = *blockchain.CreateGenesis(newBlock.TimeStamp)
			(*block_chain).Blocks[0].Data = newBlock.Data
		} else {
			(*block_chain).AddNewblock(newBlock.Data, newBlock.TimeStamp)
		}

	}
	return nil
}
