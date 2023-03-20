package blockchain

type BlockChain struct {
	Blocks []*Block
}

func (c *BlockChain) GenesisExists() bool {
	return string(c.Blocks[0].Data) != ""
}

func InitBlockChain(time string) *BlockChain {
	return &BlockChain{Blocks: []*Block{CreateGenesis(time)}}
}

func (c *BlockChain) AddNewblock(data []byte, time string) {
	prevBlock := c.Blocks[len(c.Blocks)-1]
	n := CreateBlock(data, prevBlock.Hash, time)
	c.Blocks = append(c.Blocks, n)
}

func (b *BlockChain) PrintChain() {
	for _, d := range b.Blocks {
		d.PrintBlock()
	}
}
