package hoji

// Blockchain is
type Blockchain struct {
	Blocks []*Block
}

// NewBlockchain creates and returns an instance of the Blockchain struct
func NewBlockchain() *Blockchain {
	return &Blockchain{
		[]*Block{NewGenesisBlock()},
	}
}

//AddBlock adds a new block to the blockchain
func (bc *Blockchain) AddBlock(data []byte) {
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.Blocks = append(bc.Blocks, newBlock)
}
