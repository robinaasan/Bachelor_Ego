package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

// type Transaction struct {
// 	From []byte
// 	To []byte
// }

type Block struct {
	TimeStamp string `json:"TimeStamp"`
	Hash      []byte `json:"Hash"`
	Data      string `json:"Data"` //will be a transaction
	PrevHash  []byte `json:"PrevHash"`
}

func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{[]byte(b.Data), b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}

func (b *Block) Serialize() ([]byte, error) {
	jsonBody, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return jsonBody, nil
}

func CreateBlock(data string, prevHash []byte, time string) *Block {
	block := &Block{TimeStamp: time, Hash: []byte{}, Data: data, PrevHash: prevHash}
	block.DeriveHash()
	return block
}

func CreateGenesis(time string) *Block {
	return CreateBlock("Genesis", []byte{}, time)
}

func (b *Block) PrintBlock() {
	//const layout = "2006-01-02 15:04:05.999999999 -0700 MST"
	//timeStamp, _ := time.Parse(layout, b.TimeStamp)
	fmt.Printf("Timestamp %s\n", b.TimeStamp)
	fmt.Printf("Prev hash: %x\n", b.PrevHash)
	fmt.Printf("Data: %s\n", b.Data)
	fmt.Printf("Hash: %x\n", b.Hash)
	fmt.Println()
}

// func SetGenesysFile(filename string) {
// 	file, err := os.ReadFile(filename)
// 	if err != nil {
// 		panic(err)
// 	}
// 	json.Unmarshal(file, &block_chain)
// 	timeNow := time.Now()
// 	genesysBlock := &Block{
// 		OldTransaction: Transaction{TimeStamp: timeNow},
// 		NewTransaction: Transaction{TimeStamp: timeNow},
// 		Hash:           "genesys",
// 	}
// 	block_chain = append(block_chain, *genesysBlock)

// 	fmt.Printf("%+v", block_chain)

// 	// Preparing the data to be marshalled and written.
// 	dataBytes, err := json.Marshal(block_chain)
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = os.WriteFile(filename, dataBytes, 0644)
// 	if err != nil {
// 		panic(err)
// 	}
// }
