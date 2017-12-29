package hoji

import (
	"encoding/hex"
	"fmt"

	"github.com/boltdb/bolt"
)

const utxoBucket = "chainstate"

//UTXOSet is
type UTXOSet struct {
	Bc *Blockchain
}

//SpendableOutput is
type SpendableOutput struct {
	TxID  []byte
	Value int
	index int
}

//CreateUTXOSet is
func CreateUTXOSet(bc *Blockchain) error {
	if err := bc.DB.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(utxoBucket)); err != nil && err != bolt.ErrBucketNotFound {
			return fmt.Errorf("error deleting bucket %v", err)
		}
		if _, err := tx.CreateBucket([]byte(utxoBucket)); err != nil {
			return fmt.Errorf("error creating bucket %v", err)
		}
		return nil
	}); err != nil {
		return err
	}

	utxo, err := bc.ListUTXO()
	if err != nil {
		return err
	}

	return bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for txID, outputs := range utxo {
			key, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}

			serialized, err := outputs.Bytes()
			if err != nil {
				return err
			}

			err = b.Put(key, serialized)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// FindUTXO finds UTXO for an address
func (u UTXOSet) FindUTXO(address []byte) ([]*TxOutput, error) {
	var UTXOs []*TxOutput

	pubKeyHash := ExtractPubKeyHash(address)

	if err := u.Bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs, err := BytesToOutputs(v)
			if err != nil {
				return err
			}

			for _, out := range outs.Outputs {
				fmt.Println("hash", hex.EncodeToString(out.PubKeyHash))
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return UTXOs, nil
}

//FindSpendableOutputs is
func (u *UTXOSet) FindSpendableOutputs(address []byte) ([]*SpendableOutput, error) {
	var spendableOutput []*SpendableOutput

	pubKeyHash := ExtractPubKeyHash(address)

	if err := u.Bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs, err := BytesToOutputs(v)
			if err != nil {
				return err
			}

			for i, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					so := &SpendableOutput{
						TxID:  k,
						Value: out.Value,
						index: i,
					}
					spendableOutput = append(spendableOutput, so)
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return spendableOutput, nil
}

//Update is
func (u *UTXOSet) Update(block *Block) error {
	return u.Bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if !tx.IsCoinbase() {
				for _, input := range tx.Inputs {
					updatedOuts := TxOutputs{}
					outsBytes := b.Get(input.TxID)
					outs, err := BytesToOutputs(outsBytes)
					if err != nil {
						return err
					}

					for outIndx, out := range outs.Outputs {
						if outIndx != input.outIndex {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}

						if len(updatedOuts.Outputs) == 0 {
							if err := b.Delete(input.TxID); err != nil {
								return err
							}
						} else {
							s, err := updatedOuts.Bytes()
							if err != nil {
								return err
							}
							if err := b.Put(input.TxID, s); err != nil {
								return err
							}
						}
					}
				}
			}

			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			s, err := newOutputs.Bytes()
			if err != nil {
				return err
			}
			if err := b.Put(tx.ID, s); err != nil {
				return err
			}
		}

		return nil
	})
}
