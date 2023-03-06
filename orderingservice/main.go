package main

import (
	"fmt"
	"net/http"

	"github.com/robinaasan/Bachelor_Ego/orderingservice/blockchain"
)

// MyStruct is an example structure for this program.

// var genesys_created = false
// var block_chain []Block
var block_chain *blockchain.BlockChain

func main() {
	// if !genesys_created {
	// 	filename := "genesys.json"
	// 	err := checkFile(filename)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	//updateBlockchain(filename, "genesys", "OK")
	// 	block_chain = newBlockChain()
	// 	setGenesysBlock(filename)
	// }
	//TODO: verify the integrity of the blocks if there is a genesis block
	if !block_chain.GenesisExists() {
		block_chain = blockchain.InitBlockChain()
		//create initblock file
	}

	http.HandleFunc("/", handlerTransaction)

	server := http.Server{Addr: "localhost:8083"}
	fmt.Println("Listening...")
	err := server.ListenAndServe()
	fmt.Println(err)

}

func handlerTransaction(w http.ResponseWriter, r *http.Request) {

	//fmt.Println("Funker med client")

}

// func checkFile(filename string) error {
// 	_, err := os.Stat(filename)
// 	if os.IsNotExist(err) {
// 		_, err := os.Create(filename)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func updateBlockchain(filename string, key string, val string) {

// 	file, err := os.ReadFile(filename)
// 	if err != nil {
// 		panic(err)
// 	}

// 	//data := []Transaction{}

// 	// Here the magic happens!
// 	json.Unmarshal(file, &block_chain)

// 	newStruct := &Block{
// 		OldTransaction: Transaction{Key: key, Val: val},
// 	    NewTransaction: Transaction{Key: key, Val: val},
// 	}

// 	data = append(data, *newStruct)
// 	fmt.Printf("%+v", data)

// 	// Preparing the data to be marshalled and written.
// 	dataBytes, err := json.Marshal(data)
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = ioutil.WriteFile(filename, dataBytes, 0644)
// 	if err != nil {
// 		panic(err)
// 	}
// }
