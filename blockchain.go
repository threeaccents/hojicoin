package hoji

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
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
		if err := CreateBlockchain([]byte("genesis address")); err != nil {
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
func CreateBlockchain(address []byte) error {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	transaction, err := NewCoinbaseTx(address, []byte("yolo dolo"))
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
func (bc *Blockchain) FindUnspentTxs(address []byte) ([]*Transaction, error) {
	pubKeyHash := ExtractPubKey(address)
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
					fmt.Println("output was spent")
					for _, spentOutput := range spentTxOutputs[txID] {
						if spentOutput == outTxIndex {
							continue Outputs
						}
					}
				}

				if outTx.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, tx)
				}
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					ok, err := in.UsesKey(pubKeyHash)
					if err != nil {
						return nil, err
					}
					if ok {
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

	return unspentTxs, nil
}

//FindUTXO finds all unspent transaction outputs
func (bc *Blockchain) FindUTXO(address []byte) ([]*TxOutput, error) {
	var outputs []*TxOutput

	txs, err := bc.FindUnspentTxs(address)
	if err != nil {
		return nil, err
	}

	pubKey := ExtractPubKey(address)

	for _, tx := range txs {
		for _, output := range tx.Outputs {
			if output.IsLockedWithKey(pubKey) {
				outputs = append(outputs, output)
			}
		}
	}

	return outputs, nil
}

//FindTx is
func (bc *Blockchain) FindTx(id []byte) (*Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, id) == 0 {
				return tx, nil
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return nil, ErrNotFound
}

//SignTx is
func (bc *Blockchain) SignTx(tx *Transaction, privKey *ecdsa.PrivateKey) error {
	prevTxs := make(map[string]*Transaction)

	for _, input := range tx.Inputs {
		prevTx, err := bc.FindTx(input.TxID)
		if err != nil {
			return err
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	return tx.Sign(privKey, prevTxs)
}

//VerifyTransaction is
func (bc *Blockchain) VerifyTransaction(tx *Transaction) (bool, error) {
	prevTxs := make(map[string]*Transaction)

	for _, input := range tx.Inputs {
		prevTx, err := bc.FindTx(input.TxID)
		if err != nil {
			return false, err
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	return tx.Verify(prevTxs)
}

//MineBlock adds a new block to the blockchain
func (bc *Blockchain) MineBlock(txs []*Transaction) error {
	for _, tx := range txs {
		ok, err := bc.VerifyTransaction(tx)
		if err != nil {
			return err
		}

		if !ok {
			return errors.New("invalid transaction")
		}
	}

	var lastHash []byte
	if err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte(lastHashKey))
		return nil
	}); err != nil {
		return err
	}

	newBlock := NewBlock(txs, lastHash)
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
