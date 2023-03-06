package blockchain

import (
	"bytes"
	"crypto/sha256"
)

// type Transaction struct {
// 	From []byte
// 	To []byte
// }

type Block struct {
	Hash     []byte
	Data     []byte //will be a transaction
	PrevHash []byte
}

func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}

func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{Hash: []byte{}, Data: []byte(data), PrevHash: prevHash}
	block.DeriveHash()
	return block
}

func CreateGenesis() *Block {
	return CreateBlock("Genesis", []byte{})
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
