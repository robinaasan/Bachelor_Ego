package blockchain

type BlockChain struct {
	blocks []*Block
}

func (c *BlockChain) GenesisExists() bool {
	return c.blocks[0].Data != nil
}

func InitBlockChain() *BlockChain {
	return &BlockChain{blocks: []*Block{CreateGenesis()}}
}

func (c *BlockChain) AddNewblock(data string) {
	prevBlock := c.blocks[len(c.blocks)-1]
	n := CreateBlock(data, prevBlock.PrevHash)
	c.blocks = append(c.blocks, n)
}
