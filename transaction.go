package hoji

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
)

// Transaction represents a Hoji transaction. Maybe split transaction into 2 separe structs transaction and coinbase transaction
type Transaction struct {
	ID      []byte
	Inputs  []*TxInput
	Outputs []*TxOutput
}

// subsidy is the reward "coins" given for a miner
const subsidy = 10

// NewCoinbaseTx a coinbase transaction is a transaction that does not require inputs to generate outputs. The gensis block is a coinbase transaction and when miners mine new blocks their reward is a coinbase transaction.
func NewCoinbaseTx(to, data string) (*Transaction, error) {
	txIn := &TxInput{
		ScriptSig: data,
		outIndex:  -1,
	}
	txOut := &TxOutput{
		Value:        subsidy,
		ScriptPubKey: to,
	}

	tx := &Transaction{
		Inputs:  []*TxInput{txIn},
		Outputs: []*TxOutput{txOut},
	}
	txID, err := tx.hashTransaction()
	if err != nil {
		return nil, err
	}
	tx.ID = txID
	return tx, nil
}

//NewTx is
func (bc *Blockchain) NewTx(from, to string, amount int) (*Transaction, error) {
	var inputs []*TxInput
	var outputs []*TxOutput

	txs := bc.FindUnspentTxs(from)

	accumulated := 0
Work:
	for _, tx := range txs {
		for i, output := range tx.Outputs {
			if output.CanBeUnlockedWith(from) && accumulated < amount {
				accumulated += output.Value
				input := &TxInput{
					TxID:      tx.ID,
					outIndex:  i,
					ScriptSig: from,
				}
				inputs = append(inputs, input)
				if accumulated > amount {
					break Work
				}
			}

		}
	}

	if accumulated < amount {
		return nil, ErrInsuficientFunds
	}

	txOutput := &TxOutput{
		Value:        amount,
		ScriptPubKey: to,
	}
	outputs = append(outputs, txOutput)
	if accumulated > amount {
		change := accumulated - amount
		changeOutput := &TxOutput{
			Value:        change,
			ScriptPubKey: from,
		}
		outputs = append(outputs, changeOutput)
	}
	tx := &Transaction{
		Outputs: outputs,
		Inputs:  inputs,
	}

	txID, err := tx.hashTransaction()
	if err != nil {
		return nil, err
	}
	tx.ID = txID

	return tx, nil
}

//hashTransaction will hash all the transactions contents using sha256. hashTransaction will transform the transaction struct pointer into a byte array then sha256 hash it returing the hash.
func (t *Transaction) hashTransaction() ([]byte, error) {
	var txBytes bytes.Buffer
	if err := gob.NewEncoder(&txBytes).Encode(t); err != nil {
		return nil, err
	}

	hash := sha256.Sum256(txBytes.Bytes())
	return hash[:], nil
}

// IsCoinbase checks whether the transaction is a coinbase tx
func (t *Transaction) IsCoinbase() bool {
	return len(t.Inputs) == 1 && len(t.Inputs[0].TxID) == 0 && t.Inputs[0].outIndex == -1
}
