package hoji

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"time"
)

// Block is the data structure that holds the blockchain's data.In bitcoin the block holds an array on transactions. Their block size limit is 1mb.
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

//NewBlock creates and returns a new block for the blockchain
func NewBlock(tx []*Transaction, PrevBlockHash []byte) *Block {
	b := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  tx,
		PrevBlockHash: PrevBlockHash,
	}
	// We need the other properties of the Block to be set to generate a hash. That's why we have a special method for it that we call after setting the value for the other Block struct properties
	b.SetHash()

	return b
}

//NewGenesisBlock creates and returns a new genesis block for the blockchain. The genesis block is the first block created in the blockchain. Since a block needs a previous block to be created we much create the first block "artificially"
func NewGenesisBlock(coinbaseTx *Transaction) *Block {
	return NewBlock([]*Transaction{coinbaseTx}, []byte{})
}

//SetHash creates the hash(I like to think of it as the block's ID) for a block.
// NOTE: should I just return a new block? It is more computationally expensive but makes for better code debugging imo.
func (b *Block) SetHash() {
	pow := NewPOW(b)
	hash, nonce := pow.Exec()

	b.Hash = hash
	b.Nonce = nonce
}

//HashTransactions will hash the blocks transaction struct slice and return it. It does this by concatinating all the transaction ids and then sha256 hasing them.
func (b *Block) HashTransactions() []byte {
	var txIds [][]byte
	for _, tx := range b.Transactions {
		txIds = append(txIds, tx.ID)
	}

	hash := sha256.Sum256(bytes.Join(txIds, []byte{}))
	return hash[:]
}

//Bytes transforms a Block struct to a byte array
func (b *Block) Bytes() ([]byte, error) {
	result := new(bytes.Buffer)
	if err := gob.NewEncoder(result).Encode(b); err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

//BytesToBlock tranforms a byte array into a Block struct
func BytesToBlock(v []byte) (*Block, error) {
	b := new(Block)
	if err := gob.NewDecoder(bytes.NewReader(v)).Decode(b); err != nil {
		return nil, err
	}
	return b, nil
}
