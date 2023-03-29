package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edgelesssys/ego/enclave"
	"github.com/robinaasan/Bachelor_Ego/orderingservice/blockchain"
)

const (
	PATH    = "./blockFiles/"
	genesis = "Block1.json"
)

var runtimes = []string{"http://localhost:8086/Callback", "http://localhost:8090/Callback"}

// name below should be replaces by som hash later
type Transaction struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	OldVal     int    `json:"OldVal"`
	ClientName string `json:"ClientName"`
}

type ResponsesRuntime struct {
	response string
	endpoint string
	err      error
}

type BlockTransactionStore struct {
	wg sync.WaitGroup
	sync.Mutex
	allTransactions []*Transaction
	blockchain      *blockchain.BlockChain
	count           int
}

func main() {
	// TODO: verify the integrity of the blocks if there is a genesis block
	rep, err := enclave.GetSelfReport()
	if err != nil {
		panic(err)
	}
	signature_enc := rep.SignerID
	genBlock := fmt.Sprintf("%s%s", PATH, genesis)
	block_chain := blockchain.InitBlockChain(time.Now().String(), signature_enc)
	blockSice := 5
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
	blockTransactionStore.blockchain.PrintChain()
	http.HandleFunc("/", blockTransactionStore.handlerTransaction(blockSice, signature_enc))

	server := http.Server{Addr: "localhost:8087"}
	fmt.Println("Listening...")
	err = server.ListenAndServe()
	fmt.Println(err)
}

// Add the block to the blockChain
// TODO: notify the runtimes about the change!
func (bt *BlockTransactionStore) handlerTransaction(blockSice int, signature_enc []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//start := time.Now()
		newTransAction := &Transaction{}
		err := json.NewDecoder(r.Body).Decode(newTransAction)
		if err != nil {
			fmt.Fprintf(w, "Error reading the transaction")
			return
		}
		// fmt.Printf("%+v", newTransAction)
		if err != nil {
			fmt.Fprintf(w, "Error transforming the transaction")
			return
		}
		bt.Lock()
		defer bt.Unlock()

		bt.allTransactions = append(bt.allTransactions, newTransAction)
		bt.count++
		if bt.count == blockSice {

			bt.count = 0
			allTransactionBytes, err := json.Marshal(bt.allTransactions)
			if err != nil {
				fmt.Fprintf(w, "Error: decoding the transaction went wrong")
				return
			}
			// block_chain.AddNewblock(transactionData, time.Now().String(), clientName)
			bt.blockchain.AddNewblock(allTransactionBytes, signature_enc, time.Now().String())
			addedBlock := bt.blockchain.Blocks[len(bt.blockchain.Blocks)-1]
			newBlockFileName := fmt.Sprintf("%s%s.json", PATH, fmt.Sprintf("Block%v", len(bt.blockchain.Blocks)))
			// fmt.Printf("%x\n", addedBlock.Hash)
			// fmt.Println(newBlockFileName)
			err = addBlockFile(newBlockFileName, addedBlock)
			if err != nil {
				fmt.Fprintf(w, "Error adding the block in the blockchain")
				return
			}

			// responselist := make([]ResponsesRuntime, 1)
			cl := &http.Client{}
			bt.sendCallback(allTransactionBytes, runtimes, cl)
			// if err != nil {
			// 	fmt.Printf("Error: %v", err)
			// }

			// new rquest to every runtime connected with x new transactions
			bt.allTransactions = nil

		}
		//fmt.Printf("%v ms elapsed\n", time.Since(start).Microseconds())
		// fmt.Printf("%.4fms elapsed", time.Since(start).Milliseconds())
		// fmt.Fprintf(w, "ACK")
		// s := fmt.Sprintf("%s", r.RemoteAddr)
	}
}

func (bt *BlockTransactionStore) sendCallback(allTransactionBytes []byte, endpoints []string, cl *http.Client) {
	// var wg sync.WaitGroup
	c := make(chan ResponsesRuntime)
	for _, endpoint := range endpoints {
		bt.wg.Add(1)
		go checkURL(endpoint, c, &bt.wg, allTransactionBytes, cl)
	}
	go func() {
		bt.wg.Wait()
		close(c)
	}()

	// for r := range c {
	// 	// if r.err != nil {

	// 	// 	s := fmt.Sprintf("Error: endpoint: %s got: %v\n", r.endpoint, r.err)
	// 	// 	fmt.Printf("%v", s)
	// 	// } else {
	// 	// 	fmt.Println(r.response + "\n")
	// 	// }

	// 	// if r.err != nil {
	// 	// 	fmt.Printf("Error requesting %s: %v\n", r.endpoint, r.err)
	// 	// 	continue
	// 	// }
	// 	fmt.Printf("%+v\n", r)
	// }
}

func checkURL(endpoint string, c chan ResponsesRuntime, wg *sync.WaitGroup, allTransactionBytes []byte, cl *http.Client) {
	defer (*wg).Done()

	// responseruntime := ResponsesRuntime{endpoint: endpoint}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(allTransactionBytes))
	// if err != nil {
	// 	s = err.Error()
	// }
	if err != nil {
		c <- ResponsesRuntime{endpoint, "", err}
		return
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := cl.Do(req)
	if err != nil {
		c <- ResponsesRuntime{endpoint, "", err}
		return
	}
	defer res.Body.Close()
	// resBody, err := io.ReadAll(res.Body)

	// fmt.Printf("Res: %v", string(resBody))
	c <- ResponsesRuntime{endpoint, res.Status, nil}
}

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
			(*blockTransactionStore).blockchain.Blocks[0] = blockchain.CreateGenesis(newBlock.TimeStamp, newBlock.SignID)
			(*blockTransactionStore).blockchain.Blocks[0].Data = newBlock.Data
		} else { // genesis block is already created in the filesystem
			(*blockTransactionStore).blockchain.AddNewblock(newBlock.Data, newBlock.SignID, newBlock.TimeStamp)
		}

	}
	return nil
}
