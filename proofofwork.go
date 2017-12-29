package hoji

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64
)

// targetBits is how complicated we want to make our hashcash proof. For our example we are saying the first 24 bits or 8 bytes or 3 characters of the hash must be 0
const targetBits = 24

//ProofOfWork is
type ProofOfWork struct {
	Block  *Block
	target *big.Int
}

//NewPOW is
func NewPOW(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	return &ProofOfWork{
		Block:  b,
		target: target,
	}
}

// Exec executes the pow for a new block
func (p *ProofOfWork) Exec() ([]byte, int) {
	var hashInt big.Int
	nonce := 0

	fmt.Println("Mining new block")
	for {
		if nonce > maxNonce {
			log.Panic("proof of work nonce overflow")
		}

		preppedData := p.prepData(nonce)
		hash := sha256.Sum256(preppedData)
		hashInt.SetBytes(hash[:])
		fmt.Printf("\rtesting hash: %x : hash number: %s", hash, hashInt.String())

		if hashInt.Cmp(p.target) == -1 {
			fmt.Print("\n\n")
			return hash[:], nonce
		}
		nonce++
	}
}

//Validate validates if a hash has met its requirments
func (p *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := p.prepData(p.Block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(p.target) == -1
}

//prepData will convert all of the pow data into bytes.
func (p *ProofOfWork) prepData(nonce int) []byte {
	return bytes.Join(
		[][]byte{
			p.Block.PrevBlockHash,
			IntToByte(p.Block.Timestamp),
			p.Block.HashTransactions(),
			IntToByte(int64(targetBits)),
			IntToByte(int64(nonce)),
		},
		[]byte{},
	)
}
