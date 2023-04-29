package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type Block struct {
	TimeStamp string `json:"TimeStamp"`
	Hash      []byte `json:"Hash"`
	Data      []byte `json:"Data"` //will be a transaction
	PrevHash  []byte `json:"PrevHash"`
	SignID    []byte `json:"Signature"`
}

func (b *Block) DeriveHash() {
	b.Hash = calculateHash(b)
}

func calculateHash(block *Block) []byte {
    info := bytes.Join([][]byte{block.Data, block.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
    return hash[:]
}

func (b *Block) Serialize() ([]byte, error) {
	jsonBody, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	return jsonBody, nil
}

func CreateBlock(data []byte, prevHash []byte, time string, signID []byte) *Block {
	block := &Block{TimeStamp: time, Hash: []byte{}, Data: data, PrevHash: prevHash, SignID: signID}
	block.DeriveHash()
	return block
}

func CreateGenesis(time string, signID []byte) *Block {
	return CreateBlock([]byte("Genesis"), []byte{}, time, signID)
}

func (b *Block) PrintBlock() {
	fmt.Printf("Timestamp %s\n", b.TimeStamp)
	fmt.Printf("Prev hash: %x\n", b.PrevHash)
	fmt.Printf("Data: %x\n", b.Data)
	fmt.Printf("Hash: %x\n", b.Hash)
	fmt.Printf("Signature: %x\n", b.SignID)
	fmt.Println()
}
