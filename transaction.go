package hoji

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"log"
	"math/big"
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
func NewCoinbaseTx(to, data []byte) (*Transaction, error) {
	if data == nil {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			return nil, err
		}

		data = randData
	}
	txIn := &TxInput{
		TxID:      []byte{},
		PubKey:    data,
		Signature: nil,
		outIndex:  -1,
	}
	txOut := NewTxOutput(subsidy, to)

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
func (bc *Blockchain) NewTx(from, to []byte, amount int) (*Transaction, error) {
	var inputs []*TxInput
	var outputs []*TxOutput

	wallets, err := NewWallets()
	if err != nil {
		return nil, err
	}
	wallet := wallets.GetWallet(string(from))

	utxoSet := UTXOSet{Bc: bc}

	spendableOutputs, err := utxoSet.FindSpendableOutputs(from)
	if err != nil {
		return nil, err
	}

	accumulated := 0
	for _, so := range spendableOutputs {
		accumulated += so.Value
		in := &TxInput{
			TxID:     so.TxID,
			outIndex: so.index,
			PubKey:   wallet.PublicKey,
		}
		inputs = append(inputs, in)
	}

	outputs = append(outputs, NewTxOutput(amount, to))
	if accumulated > amount {
		change := accumulated - amount
		outputs = append(outputs, NewTxOutput(change, from)) // a change
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

	if err := bc.SignTx(tx, wallet.PrivateKey); err != nil {
		return nil, err
	}

	return tx, nil
}

//Sign is
func (t *Transaction) Sign(privateKey *ecdsa.PrivateKey, prevTxs map[string]*Transaction) error {
	if t.IsCoinbase() {
		return nil
	}

	for _, input := range t.Inputs {
		if prevTxs[hex.EncodeToString(input.TxID)].ID == nil {
			return errors.New("ERROR: Previous transaction is not correct")
		}
	}

	trimmedTx := t.Trim()

	for inputIndex, input := range trimmedTx.Inputs {
		prevTx := prevTxs[hex.EncodeToString(input.TxID)]
		trimmedTx.Inputs[inputIndex].Signature = nil
		trimmedTx.Inputs[inputIndex].PubKey = prevTx.Outputs[input.outIndex].PubKeyHash
		x := new(big.Int)
		y := new(big.Int)
		keyLen := len(input.PubKey)
		x.SetBytes(trimmedTx.Inputs[inputIndex].PubKey[:(keyLen / 2)])
		y.SetBytes(trimmedTx.Inputs[inputIndex].PubKey[(keyLen / 2):])

		txID, err := trimmedTx.hashTransaction()
		if err != nil {
			return err
		}
		trimmedTx.ID = txID
		trimmedTx.Inputs[inputIndex].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, privateKey, trimmedTx.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		t.Inputs[inputIndex].Signature = signature

	}

	return nil
}

//Verify is
func (t *Transaction) Verify(prevTxs map[string]*Transaction) (bool, error) {
	if t.IsCoinbase() {
		return true, nil
	}

	for _, input := range t.Inputs {
		if prevTxs[hex.EncodeToString(input.TxID)].ID == nil {
			return false, errors.New("ERROR: Previous transaction is not correct")
		}
	}

	trimmedTx := t.Trim()
	curve := elliptic.P256()

	for inputIndex, input := range t.Inputs {
		prevTx := prevTxs[hex.EncodeToString(input.TxID)]
		trimmedTx.Inputs[inputIndex].Signature = nil
		trimmedTx.Inputs[inputIndex].PubKey = prevTx.Outputs[input.outIndex].PubKeyHash
		txID, err := trimmedTx.hashTransaction()
		if err != nil {
			return false, err
		}
		trimmedTx.ID = txID
		trimmedTx.Inputs[inputIndex].PubKey = nil

		r := new(big.Int)
		s := new(big.Int)
		sigLen := len(input.Signature)
		r.SetBytes(input.Signature[:(sigLen / 2)])
		s.SetBytes(input.Signature[(sigLen / 2):])

		x := new(big.Int)
		y := new(big.Int)
		keyLen := len(input.PubKey)
		x.SetBytes(input.PubKey[:(keyLen / 2)])
		y.SetBytes(input.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		}
		if ecdsa.Verify(&rawPubKey, trimmedTx.ID, r, s) == false {
			return false, nil
		}
	}

	return true, nil
}

//Trim is
func (t *Transaction) Trim() *Transaction {
	var inputs []*TxInput
	var outputs []*TxOutput

	for _, input := range t.Inputs {
		in := &TxInput{
			TxID:     input.TxID,
			outIndex: input.outIndex,
		}
		inputs = append(inputs, in)
	}

	for _, ouput := range outputs {
		out := &TxOutput{
			Value:      ouput.Value,
			PubKeyHash: ouput.PubKeyHash,
		}
		outputs = append(outputs, out)
	}

	return &Transaction{
		ID:      t.ID,
		Inputs:  inputs,
		Outputs: outputs,
	}
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
