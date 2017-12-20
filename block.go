package hoji

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

// Block is the data structure that holds the blockchain's data.In bitcoin the block holds an array on transactions. Their block size limit is 1mb.
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
}

//NewBlock creates and returns a new block for the blockchain
func NewBlock(data, PrevBlockHash []byte) *Block {
	b := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          data,
		PrevBlockHash: PrevBlockHash,
	}
	// We need the other properties of the Block to be set to generate a hash. That's why we have a special method for it that we call after setting the value for the other Block struct properties
	b.SetHash()

	return b
}

//NewGenesisBlock creates and returns a new genesis block for the blockchain. The genesis block is the first block created in the blockchain. Since a block needs a previous block to be created we much create the first block "artificially"
func NewGenesisBlock() *Block {
	return NewBlock([]byte("hoji coin genesis block"), []byte{})
}

//SetHash creates the hash(I like to think of it as the blocks ID) for a block.
// NOTE: should I just return a new block? It is more computationally expensive but makes for better code debugging imo.
func (b *Block) SetHash() {
	strTimestamp := strconv.FormatInt(b.Timestamp, 10)
	timestamp := []byte(strTimestamp)

	//This is all the Block struct properties combined into one byte array
	headers := bytes.Join(
		[][]byte{
			b.Data,
			b.PrevBlockHash,
			timestamp,
		},
		[]byte{},
	)

	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}
