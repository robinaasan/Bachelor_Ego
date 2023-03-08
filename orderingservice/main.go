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

// MyStruct is an example structure for this program.

// var genesys_created = false
// var block_chain []Block
var block_chain *blockchain.BlockChain

const PATH = "./blockFiles/"
const GenesisFile = "genesys.json"

//const DIR = "./blockFiles"

type Transaction struct {
	Key    int `json:"Key"`
	NewVal int `json:"NewVal"`
	OldVal int `json:"OldVal"`
}

func main() {
	//TODO: verify the integrity of the blocks if there is a genesis block
	genBlock := fmt.Sprintf("%s%s", PATH, GenesisFile)
	block_chain = blockchain.InitBlockChain(time.Now().String())
	if !fileExist(genBlock) {
		//TODO: create on with the genesys block

		//TODO: create the genesis block

		//Create the file with the genesis block

		err := addBlockFile(genBlock, block_chain.Blocks[0])
		if err != nil {
			fmt.Println(err)
		}

	} else {
		//TODO: load rest of blockchain
		err := ReadAllBlockFiles()
		if err != nil {
			fmt.Println(err)
		}
		//block_chain.PrintChain()
	}
	http.HandleFunc("/", handlerTransaction)

	server := http.Server{Addr: "localhost:8087"}
	fmt.Println("Listening...")
	err := server.ListenAndServe()
	fmt.Println(err)

}

func handlerTransaction(w http.ResponseWriter, r *http.Request) {

	//fmt.Println("Funker med client")
	newTransAction := &Transaction{}
	err := json.NewDecoder(r.Body).Decode(newTransAction)
	if err != nil {
		fmt.Fprintf(w, "Error reading the transaction")
		return
	}
	fmt.Printf("%+v", newTransAction)
	//Add to blockChain
	transactionData, err := json.Marshal(newTransAction)
	if err != nil {
		fmt.Fprintf(w, "Error transforming the transaction")
		return
	}
	block_chain.AddNewblock(transactionData, time.Now().String())

	addedBlock := block_chain.Blocks[len(block_chain.Blocks)-1]
	newBlockFileName := fmt.Sprintf("%s%x.json", PATH, addedBlock.Hash)
	fmt.Println(newBlockFileName)
	err = addBlockFile(newBlockFileName, addedBlock)

	if err != nil {
		fmt.Fprintf(w, "Error adding the block in the blockchain")
	}
	//block_chain.PrintChain()
}

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

//The genesis block is created in the main function
func ReadAllBlockFiles() error {
	files, err := os.ReadDir(PATH)
	if err != nil {
		return err
	}

	for _, file_entry := range files {
		if !file_entry.IsDir() {
			fileType := strings.Split(file_entry.Name(), ".")
			if fileType[1] != "json" {
				return errors.New("wrong file type.\n")
			}
			//if it is the genesis file create that first
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
			if fileType[0] == "genesys" {
				//block_chain = blockchain.InitBlockChain()
				block_chain.Blocks[0] = blockchain.CreateGenesis(newBlock.TimeStamp)
				block_chain.Blocks[0].Data = newBlock.Data
				//block_chain.PrintChain()
			} else { //genesis block is already created
				block_chain.AddNewblock(newBlock.Data, newBlock.TimeStamp)
			}

			//newBlock.PrintBlock()

			// newBlockBytes, err := newBlock.Serialize()

			// if err != nil {
			// 	return err
			// }
			//time, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", newBlock.TimeStamp)
			// if err != nil {
			// 	return err
			// }
		}
	}
	return nil
}
