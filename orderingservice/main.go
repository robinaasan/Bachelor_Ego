package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/robinaasan/Bachelor_Ego/orderingservice/blockchain"
)

var block_chain *blockchain.BlockChain

const PATH = "./blockFiles/"
const GenesisFile = "genesys.json"

//name below should be replaces by som hash later
type Transaction struct {
	Key        int    `json:"Key"`
	NewVal     int    `json:"NewVal"`
	OldVal     int    `json:"OldVal"`
	ClientName string `json:"ClientName"`
}

var count int
var allTransactions []byte

func main() {
	//TODO: verify the integrity of the blocks if there is a genesis block
	genBlock := fmt.Sprintf("%s%s", PATH, GenesisFile)
	block_chain = blockchain.InitBlockChain(time.Now().String())
	if !fileExist(genBlock) {
		err := addBlockFile(genBlock, block_chain.Blocks[0])
		if err != nil {
			fmt.Println(err)
		}
	} else { //Load the rest of the blockchain
		err := ReadAllBlockFiles()
		if err != nil {
			fmt.Println(err)
		}
	}
	block_chain.PrintChain()
	http.HandleFunc("/", handlerTransaction)

	server := http.Server{Addr: "localhost:8087"}
	fmt.Println("Listening...")
	err := server.ListenAndServe()
	fmt.Println(err)
}

//Add the block to the blockChain
//TODO: notify the runtimes about the change!
func handlerTransaction(w http.ResponseWriter, r *http.Request) {
	newTransAction := &Transaction{}
	err := json.NewDecoder(r.Body).Decode(newTransAction)
	if err != nil {
		fmt.Fprintf(w, "Error reading the transaction")
		return
	}
	fmt.Printf("%+v", newTransAction)
	transactionData, err := json.Marshal(newTransAction)
	if err != nil {
		fmt.Fprintf(w, "Error transforming the transaction")
		return
	}
	allTransactions = append(allTransactions, transactionData...)
	count++
	if count == 5 {
		count = 0
		//block_chain.AddNewblock(transactionData, time.Now().String(), clientName)
		block_chain.AddNewblock(allTransactions, time.Now().String())
		addedBlock := block_chain.Blocks[len(block_chain.Blocks)-1]
		newBlockFileName := fmt.Sprintf("%s%x.json", PATH, addedBlock.Hash)
		fmt.Println(newBlockFileName)
		err = addBlockFile(newBlockFileName, addedBlock)
		if err != nil {
			fmt.Fprintf(w, "Error adding the block in the blockchain")
			return
		}
		// new rquest to every runtime connected with x new transactions
		allTransactions = nil
	}
	fmt.Fprintf(w, "ACK")
	//s := fmt.Sprintf("%s", r.RemoteAddr)
}

//Add the block as a json file in the filesystem
func addBlockFile(filename string, b *blockchain.Block) error {
	jsonBody, err := b.Serialize()
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, jsonBody, 0644)
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

func ReadAllBlockFiles() error {
	files, err := os.ReadDir(PATH)
	if err != nil {
		return err
	}
	for _, file_entry := range files {
		if !file_entry.IsDir() {
			fileType := strings.Split(file_entry.Name(), ".")
			if fileType[1] != "json" {
				return errors.New("wrong file type\n.")
			}
			newBlock := &blockchain.Block{}
			fileBytes, err := os.ReadFile(PATH + file_entry.Name())
			if err != nil {
				return err
			}
			err = json.Unmarshal(fileBytes, newBlock)
			if err != nil {
				return err
			}
			//TODO: now the genesys block changes gets the the date updated i the blockchain, it is not created a new one
			//There is probably a better solution than this
			//if it is the genesis file create that first
			if fileType[0] == "genesys" {
				//The genesis block was created in main
				//Below we use the timestamp and set the same hash as is stored
				block_chain.Blocks[0] = blockchain.CreateGenesis(newBlock.TimeStamp)
				block_chain.Blocks[0].Data = newBlock.Data
			} else { //genesis block is already created in the filesystem
				block_chain.AddNewblock(newBlock.Data, newBlock.TimeStamp)
			}
		}
	}
	return nil
}
