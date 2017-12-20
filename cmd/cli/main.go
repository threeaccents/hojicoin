package main

import (
	"fmt"

	"gitlab.com/rodzzlessa24/hoji"
)

func main() {
	bc := hoji.NewBlockchain()

	bc.AddBlock([]byte("Send 1 BTC to Ivan"))
	bc.AddBlock([]byte("Send 2 more BTC to Ivan"))

	for _, block := range bc.Blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}
