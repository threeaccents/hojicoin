package hoji

import "bytes"

// TxInput references a previous output: Txid stores the ID of such transaction, and Vout stores an index of the  output it refrences in the transaction. ScriptSig is a script which provides data to be used in an output’s ScriptPubKey
type TxInput struct {
	TxID      []byte // the output tx it refrences
	outIndex  int
	Signature []byte
	PubKey    []byte
}

//UsesKey is
func (in *TxInput) UsesKey(pubKeyHash []byte) (bool, error) {
	lockingHash, err := hashPubKey(in.PubKey)
	if err != nil {
		return false, err
	}
	ok := bytes.Compare(lockingHash, pubKeyHash) == 0
	return ok, nil
}
