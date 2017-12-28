package hoji

// TxInput references a previous output: Txid stores the ID of such transaction, and Vout stores an index of the  output it refrences in the transaction. ScriptSig is a script which provides data to be used in an output’s ScriptPubKey
type TxInput struct {
	TxID      []byte // the output tx it refrences
	outIndex  int
	ScriptSig string
}

//CanUnlockOutputWith is
func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}
