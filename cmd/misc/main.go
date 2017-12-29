package main

import (
	"encoding/hex"
	"fmt"

	"gitlab.com/rodzzlessa24/hoji"
)

// 17f7LXxyYva2jN54N4gNhJZsiCc91VCcrh
// 4905d9d7093452122f58c8f0499cf3fd07bfba8f

// 1M6pcEFGLF2oq1nHpo519Jau15PsUJu7HP
// dc7c560c8a96a7e96c87ce0c5bfa7daa08aaa8cd

func main() {
	address := []byte("1M6pcEFGLF2oq1nHpo519Jau15PsUJu7HP")

	pubKey := hoji.ExtractPubKeyHash(address)

	fmt.Println("key: ", hex.EncodeToString(pubKey))
}
