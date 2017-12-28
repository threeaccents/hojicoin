package hoji

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
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
	txIn := &TxInput{
		PubKey:   data,
		outIndex: -1,
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
	pubKeyHash, err := hashPubKey(wallet.PublicKey)
	if err != nil {
		return nil, err
	}

	txs, err := bc.FindUnspentTxs(from)
	if err != nil {
		return nil, err
	}

	accumulated := 0
Work:
	for _, tx := range txs {
		for i, output := range tx.Outputs {
			if output.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += output.Value

				in := &TxInput{
					TxID:     tx.ID,
					outIndex: i,
					PubKey:   wallet.PublicKey,
				}

				inputs = append(inputs, in)

				if accumulated > amount {
					break Work
				}
			}

		}
	}

	if accumulated < amount {
		return nil, ErrInsuficientFunds
	}
	outputs = append(outputs, NewTxOutput(amount, to))
	if accumulated > amount {
		change := accumulated - amount
		outputs = append(outputs, NewTxOutput(change, from))
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

	trimmedTx := t.Trim()

	for inputIndex, input := range trimmedTx.Inputs {
		prevTx := prevTxs[hex.EncodeToString(input.TxID)]
		trimmedTx.Inputs[inputIndex].PubKey = prevTx.Outputs[input.outIndex].PubKeyHash
		id, err := trimmedTx.hashTransaction()
		if err != nil {
			return err
		}
		trimmedTx.ID = id
		trimmedTx.Inputs[inputIndex].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, privateKey, trimmedTx.ID)
		if err != nil {
			return err
		}

		signature := append(r.Bytes(), s.Bytes()...)
		t.Inputs[inputIndex].Signature = signature
	}

	return nil
}

//Verify is
func (t *Transaction) Verify(prevTxs map[string]*Transaction) (bool, error) {
	trimmedTx := t.Trim()
	curve := elliptic.P256()

	for inputIndex, input := range t.Inputs {
		prevTx := prevTxs[hex.EncodeToString(input.TxID)]
		trimmedTx.Inputs[inputIndex].Signature = nil
		trimmedTx.Inputs[inputIndex].PubKey = prevTx.Outputs[input.outIndex].PubKeyHash
		id, err := trimmedTx.hashTransaction()
		if err != nil {
			return false, err
		}
		trimmedTx.ID = id
		trimmedTx.Inputs[inputIndex].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		signatureLen := len(input.Signature)
		r.SetBytes(input.Signature[:(signatureLen / 2)])
		s.SetBytes(input.Signature[(signatureLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(input.PubKey)
		x.SetBytes(input.PubKey[:(keyLen / 2)])
		y.SetBytes(input.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{
			Curve: curve,
			X:     &x,
			Y:     &y,
		}
		if ecdsa.Verify(&rawPubKey, trimmedTx.ID, &r, &s) == false {
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
