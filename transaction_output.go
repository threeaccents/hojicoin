package hoji

//TxOutput is the output generated in a transaction. Outputs store "coins" in the value field. And storing means locking them with a puzzle, which is stored in the ScriptPubKey.
type TxOutput struct {
	Value        int
	ScriptPubKey string
}

//CanBeUnlockedWith is
func (out *TxOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}
