package hoji

import (
	"bytes"
	"encoding/gob"

	"gitlab.com/rodzzlessa24/hoji/base58"
)

//TxOutput is the output generated in a transaction. Outputs store "coins" in the value field. And storing means locking them with a puzzle, which is stored in the ScriptPubKey.
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

// NewTxOutput create a new TXOutput
func NewTxOutput(value int, address []byte) *TxOutput {
	txo := &TxOutput{
		Value: value,
	}
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

//TxOutputs is
type TxOutputs struct {
	Outputs []*TxOutput
}

//Bytes transforms outputs into a byte array
func (o *TxOutputs) Bytes() ([]byte, error) {
	var buff bytes.Buffer

	if err := gob.NewEncoder(&buff).Encode(o); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// BytesToOutputs deserializes TxOutputs
func BytesToOutputs(data []byte) (*TxOutputs, error) {
	var outputs *TxOutputs
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&outputs); err != nil {
		return nil, err
	}

	return outputs, nil
}
