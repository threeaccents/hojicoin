package hoji

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
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
	transaction, err := NewCoinbaseTx(address, []byte("yolo dolo"))
	if err != nil {
		return err
	}
	gensisBlock := NewGenesisBlock(transaction)
	tip := gensisBlock.Hash
	if err := db.Update(func(tx *bolt.Tx) error {
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
	}); err != nil {
		return err
	}

	return CreateUTXOSet(&Blockchain{db, tip})
}

//ListUTXO finds all unspent transaction outputs
func (bc *Blockchain) ListUTXO() (map[string]TxOutputs, error) {
	utxo := make(map[string]TxOutputs)
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

				outs := utxo[txID]
				outs.Outputs = append(outs.Outputs, outTx)
				utxo[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.TxID)
					spentTxOutputs[inTxID] = append(spentTxOutputs[inTxID], in.outIndex)
				}
			}

		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return utxo, nil
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
	if tx.IsCoinbase() {
		return true, nil
	}

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
func (bc *Blockchain) MineBlock(txs []*Transaction) (*Block, error) {
	for _, tx := range txs {
		ok, err := bc.VerifyTransaction(tx)
		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, errors.New("invalid transaction")
		}
	}

	var lastHash []byte
	if err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte(lastHashKey))
		return nil
	}); err != nil {
		return nil, err
	}

	newBlock := NewBlock(txs, lastHash)
	if err := bc.DB.Update(func(tx *bolt.Tx) error {
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
	}); err != nil {
		return nil, err
	}

	return newBlock, nil
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
