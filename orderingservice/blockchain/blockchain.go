package blockchain

type BlockChain struct {
	Blocks []*Block
}

func (c *BlockChain) GenesisExists() bool {
	return string(c.Blocks[0].Data) != ""
}

func InitBlockChain(time string, signID []byte) *BlockChain {
	return &BlockChain{Blocks: []*Block{CreateGenesis(time, signID)}}
}

func (c *BlockChain) AddNewblock(data []byte, signID []byte, time string) {
	prevBlock := c.Blocks[len(c.Blocks)-1]
	n := CreateBlock(data, prevBlock.Hash, time, signID)
	c.Blocks = append(c.Blocks, n)
}

func (b *BlockChain) PrintChain() {
	for _, d := range b.Blocks {
		d.PrintBlock()
	}
}

// Code below was a start to confirm the integrity of the blockchain but it is not finished

// func (b *BlockChain) BlockChainisNotCorrupt() bool {
// 	for i := 1; i < len(b.Blocks); i++ {
// 		b.Blocks[i].DeriveHash()

// 		// Check that the stored hash in the current block matches the calculated hash
// 		if !bytes.Equal(b.Blocks[i].Hash, calculateHash(b.Blocks[i])) {
// 			return false
// 		}

// 		// Check that the stored hash in the current block matches the hash of the previous block's data
// 		if !bytes.Equal(b.Blocks[i].PrevHash, b.Blocks[i-1].Hash) {
// 			return false
// 		}
// 	}
// 	return true
// }
