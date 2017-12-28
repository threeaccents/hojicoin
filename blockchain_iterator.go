package hoji

import (
	"log"

	"github.com/boltdb/bolt"
)

// BlockchainIterator helps us loop over all of the blocks in the blockchain
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

//Next is
func (i *BlockchainIterator) Next() *Block {
	block := new(Block)
	if err := i.db.View(func(tx *bolt.Tx) error {
		bu := tx.Bucket([]byte(blocksBucket))
		nextBlock := bu.Get(i.currentHash)
		b, err := BytesToBlock(nextBlock)
		if err != nil {
			return err
		}
		block = b
		return nil
	}); err != nil {
		log.Panic(err)
	}
	i.currentHash = block.PrevBlockHash
	return block
}
