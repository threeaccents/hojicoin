package hoji

import "github.com/boltdb/bolt"

const (
	blocksBucket = "blocks"
	lastHashKey  = "l"
)

// Blockchain is
type Blockchain struct {
	DB  *bolt.DB
	tip []byte
}

// NewBlockchain creates and returns an instance of the Blockchain struct
func NewBlockchain(db *bolt.DB) (*Blockchain, error) {
	var tip []byte
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		// If we already have a blocks bucket set the tip and return
		if b != nil {
			tip = b.Get([]byte(lastHashKey))
			return nil
		}
		gensisBlock := NewGenesisBlock()
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}

		blockBytes, err := gensisBlock.Bytes()
		if err != nil {
			return err
		}

		if err := b.Put(gensisBlock.Hash, blockBytes); err != nil {
			return err
		}

		tip = gensisBlock.Hash
		return b.Put([]byte(lastHashKey), gensisBlock.Hash)
	}); err != nil {
		return nil, err
	}

	return &Blockchain{db, tip}, nil
}

//MineBlock adds a new block to the blockchain
func (bc *Blockchain) MineBlock(data []byte) error {
	// Note: is this event needed? the blockchain struct keeps in memory the last hash. Feching from the DB seems unecessary
	// var lastHash []byte
	// if err := bc.DB.View(func(tx *bolt.Tx) error {
	// 	b := tx.Bucket([]byte(blocksBucket))
	// 	lastHash = b.Get([]byte(lastHashKey))
	// 	return nil
	// }); err != nil {
	// 	return err
	// }

	newBlock := NewBlock(data, bc.tip)
	return bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockBytes, err := newBlock.Bytes()
		if err != nil {
			return err
		}
		if err := b.Put(newBlock.Hash, blockBytes); err != nil {
			return err
		}

		if err := b.Put([]byte(lastHashKey), newBlock.Hash); err != nil {
			return err
		}
		bc.tip = newBlock.Hash
		return nil
	})
}

//Iterator returns a new iterator to loop over the blocks in the blockchain
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{
		currentHash: bc.tip,
		db:          bc.DB,
	}
}
