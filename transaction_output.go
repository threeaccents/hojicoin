package hoji

import (
	"bytes"

	"gitlab.com/rodzzlessa24/hoji/base58"
)

//TxOutput is the output generated in a transaction. Outputs store "coins" in the value field. And storing means locking them with a puzzle, which is stored in the ScriptPubKey.
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

// NewTxOutput create a new TXOutput
func NewTxOutput(value int, address []byte) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock(address)

	return txo
}

//Lock is
func (o *TxOutput) Lock(address []byte) {
	pubKeyHash := base58.Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	o.PubKeyHash = pubKeyHash
}

//IsLockedWithKey is
func (o *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(o.PubKeyHash, pubKeyHash) == 0
}
