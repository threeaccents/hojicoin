package hoji

import (
	"encoding/hex"
	"os"

	"github.com/boltdb/bolt"
)

const (
	blocksBucket = "blocks"
	lastHashKey  = "l"
	dbFile       = "hoji.db"
)

// Blockchain is
type Blockchain struct {
	DB  *bolt.DB
	tip []byte
}

// NewBlockchain creates and returns an instance of the Blockchain struct
func NewBlockchain() (*Blockchain, error) {
	if !dbExists() {
		if err := CreateBlockchain("genesis address"); err != nil {
			return nil, err
		}
	}
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	var tip []byte
	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte(lastHashKey))
		return nil
	}); err != nil {
		return nil, err
	}

	return &Blockchain{db, tip}, nil
}

//CreateBlockchain is
func CreateBlockchain(address string) error {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	transaction, err := NewCoinbaseTx(address, "yolo dolo")
	if err != nil {
		return err
	}
	gensisBlock := NewGenesisBlock(transaction)

	return db.Update(func(tx *bolt.Tx) error {
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

		return b.Put([]byte(lastHashKey), gensisBlock.Hash)
	})

}

//FindUnspentTxs is
func (bc *Blockchain) FindUnspentTxs(address string) []*Transaction {
	var unspentTxs []*Transaction
	spentTxOutputs := make(map[string][]int)
	bci := bc.Iterator()
	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outTxIndex, outTx := range tx.Outputs {
				// check if output was spent
				if spentTxOutputs[txID] != nil {
					for _, spentOutput := range spentTxOutputs[txID] {
						if spentOutput == outTxIndex {
							continue Outputs
						}
					}
				}

				if outTx.CanBeUnlockedWith(address) {
					unspentTxs = append(unspentTxs, tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.TxID)
						spentTxOutputs[inTxID] = append(spentTxOutputs[inTxID], in.outIndex)
					}
				}
			}

		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTxs
}

//FindUTXO finds all unspent transaction outputs
func (bc *Blockchain) FindUTXO(address string) []*TxOutput {
	var outputs []*TxOutput

	txs := bc.FindUnspentTxs(address)

	for _, tx := range txs {
		for _, output := range tx.Outputs {
			if output.CanBeUnlockedWith(address) {
				outputs = append(outputs, output)
			}
		}
	}

	return outputs
}

//MineBlock adds a new block to the blockchain
func (bc *Blockchain) MineBlock(tx []*Transaction) error {
	var lastHash []byte
	if err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte(lastHashKey))
		return nil
	}); err != nil {
		return err
	}

	newBlock := NewBlock(tx, lastHash)
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

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
